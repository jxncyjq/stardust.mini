package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jxncyjq/stardust.mini/utils"
)

// ApisixConfig APISIX 网关配置
type ApisixConfig struct {
	AdminURL         string `json:"admin_url" toml:"admin_url"`                   // Admin API 地址, e.g. "http://localhost:9180"
	APIKey           string `json:"api_key" toml:"api_key"`                       // Admin API Key
	Timeout          int    `json:"timeout" toml:"timeout"`                       // 请求超时(秒), 默认 5
	UpstreamAddr     string `json:"upstream_addr" toml:"upstream_addr"`           // HTTP 上游地址, e.g. "host.docker.internal:8080"
	GrpcUpstreamAddr string `json:"grpc_upstream_addr" toml:"grpc_upstream_addr"` // gRPC 上游地址, e.g. "host.docker.internal:9103"

	// RegisterMode 支持 off/single/leader。
	// off: 禁用 APISIX 写操作
	// single: 单实例直接写 APISIX（默认）
	// leader: 多实例选主后仅 Leader 写 APISIX
	RegisterMode string `json:"register_mode" toml:"register_mode"`
	// DeregisterOnShutdown 控制退出时是否删除 route/upstream。
	// nil 时按模式取默认值：leader=false，single=true。
	DeregisterOnShutdown *bool `json:"deregister_on_shutdown" toml:"deregister_on_shutdown"`

	LeaderLeaseName          string `json:"leader_lease_name" toml:"leader_lease_name"`
	LeaderLeaseNamespace     string `json:"leader_lease_namespace" toml:"leader_lease_namespace"`
	LeaderIdentity           string `json:"leader_identity" toml:"leader_identity"`
	LeaseDurationSeconds     int    `json:"lease_duration_seconds" toml:"lease_duration_seconds"`
	RenewDeadlineSeconds     int    `json:"renew_deadline_seconds" toml:"renew_deadline_seconds"`
	RetryPeriodSeconds       int    `json:"retry_period_seconds" toml:"retry_period_seconds"`
	ReconcileIntervalSeconds int    `json:"reconcile_interval_seconds" toml:"reconcile_interval_seconds"`
}

// ApisixGateway APISIX 网关实现
type ApisixGateway struct {
	config   ApisixConfig
	client   *http.Client
	mu       sync.Mutex
	routeIDs map[string][]string // serviceID -> []routeID, 用于注销时清理

	leaderMu          sync.Mutex
	leaderControllers map[string]*apisixLeaderController // serviceID -> controller
}

// NewApisixGateway 创建 APISIX 网关实例
func NewApisixGateway(apisixBytes []byte) (*ApisixGateway, error) {
	config, err := utils.Bytes2Struct[ApisixConfig](apisixBytes)
	if err != nil {
		panic(fmt.Sprintf("Apisix config error:%s", err.Error()))
	}

	if config.AdminURL == "" {
		return nil, fmt.Errorf("apisix admin_url is required")
	}
	timeout := 5
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	return &ApisixGateway{
		config:            config,
		client:            &http.Client{Timeout: time.Duration(timeout) * time.Second},
		routeIDs:          make(map[string][]string),
		leaderControllers: make(map[string]*apisixLeaderController),
	}, nil
}

func (a *ApisixGateway) GetApisixConfig() ApisixConfig {
	return a.config
}

// RegisterService 注册服务到 APISIX (创建 upstream + routes)
func (a *ApisixGateway) RegisterService(ctx context.Context, svc *GatewayService) error {
	mode := a.registerMode()
	switch mode {
	case "off":
		return nil
	case "leader":
		return a.registerServiceByLeader(ctx, svc)
	default:
		return a.registerServiceDirect(ctx, svc)
	}
}

func (a *ApisixGateway) registerServiceDirect(ctx context.Context, svc *GatewayService) error {
	// 1. 创建/更新 upstream
	upstreamID := svc.ID
	upstreamBody := toUpstreamBody(svc.Upstream, svc.Name)
	if err := a.putResource(ctx, "upstreams", upstreamID, upstreamBody); err != nil {
		return fmt.Errorf("create upstream failed: %w", err)
	}

	// 2. 创建 routes
	a.mu.Lock()
	a.routeIDs[svc.ID] = nil // 重置
	a.mu.Unlock()

	for i, route := range svc.Routes {
		routeID := fmt.Sprintf("%s-route-%d", svc.ID, i)
		routeBody := map[string]interface{}{
			"uri":  route.URI,
			"name": route.Name,
		}
		if route.Upstream != nil {
			routeBody["upstream"] = toUpstreamBody(route.Upstream, route.Name)
		} else {
			routeBody["upstream_id"] = upstreamID
		}
		if len(route.Methods) > 0 {
			routeBody["methods"] = route.Methods
		}
		if err := a.putResource(ctx, "routes", routeID, routeBody); err != nil {
			return fmt.Errorf("create route %s failed: %w", routeID, err)
		}

		a.mu.Lock()
		a.routeIDs[svc.ID] = append(a.routeIDs[svc.ID], routeID)
		a.mu.Unlock()
	}

	return nil
}

func (a *ApisixGateway) registerServiceByLeader(ctx context.Context, svc *GatewayService) error {
	a.leaderMu.Lock()
	defer a.leaderMu.Unlock()

	if old := a.leaderControllers[svc.ID]; old != nil {
		old.Stop()
	}

	controller, err := newApisixLeaderController(a, svc)
	if err != nil {
		return err
	}
	if err := controller.Start(ctx); err != nil {
		return err
	}
	a.leaderControllers[svc.ID] = controller
	return nil
}

func toUpstreamBody(upstream *GatewayUpstream, name string) map[string]interface{} {
	body := map[string]interface{}{
		"type":  upstream.Type,
		"nodes": upstream.Nodes,
		"name":  name,
	}
	if upstream.Scheme != "" {
		body["scheme"] = upstream.Scheme
	}
	return body
}

// DeregisterService 从 APISIX 注销服务 (删除 routes + upstream)
func (a *ApisixGateway) DeregisterService(ctx context.Context, serviceID string) error {
	mode := a.registerMode()
	if mode == "off" {
		return nil
	}

	if mode == "leader" {
		a.leaderMu.Lock()
		controller := a.leaderControllers[serviceID]
		delete(a.leaderControllers, serviceID)
		a.leaderMu.Unlock()
		if controller != nil {
			controller.Stop()
			if !a.shouldDeregisterOnShutdown() {
				return nil
			}
			if !controller.IsLeader() {
				return nil
			}
		}
		if !a.shouldDeregisterOnShutdown() {
			return nil
		}
	}

	if mode != "leader" && !a.shouldDeregisterOnShutdown() {
		return nil
	}

	return a.deregisterServiceDirect(ctx, serviceID)
}

func (a *ApisixGateway) deregisterServiceDirect(ctx context.Context, serviceID string) error {
	a.mu.Lock()
	routeIDs := a.routeIDs[serviceID]
	delete(a.routeIDs, serviceID)
	a.mu.Unlock()

	// 先删 routes, 再删 upstream (有依赖关系)
	for _, routeID := range routeIDs {
		if err := a.deleteResource(ctx, "routes", routeID); err != nil {
			return fmt.Errorf("delete route %s failed: %w", routeID, err)
		}
	}

	if err := a.deleteResource(ctx, "upstreams", serviceID); err != nil {
		return fmt.Errorf("delete upstream %s failed: %w", serviceID, err)
	}

	return nil
}

// Close 关闭
func (a *ApisixGateway) Close() error {
	a.leaderMu.Lock()
	controllers := make([]*apisixLeaderController, 0, len(a.leaderControllers))
	for _, controller := range a.leaderControllers {
		controllers = append(controllers, controller)
	}
	a.leaderControllers = make(map[string]*apisixLeaderController)
	a.leaderMu.Unlock()

	for _, controller := range controllers {
		controller.Stop()
	}

	a.client.CloseIdleConnections()
	return nil
}

func (a *ApisixGateway) registerMode() string {
	mode := a.config.RegisterMode
	if mode == "" {
		return "single"
	}
	switch mode {
	case "off", "single", "leader":
		return mode
	default:
		return "single"
	}
}

func (a *ApisixGateway) shouldDeregisterOnShutdown() bool {
	if a.config.DeregisterOnShutdown != nil {
		return *a.config.DeregisterOnShutdown
	}
	if a.registerMode() == "leader" {
		return false
	}
	return true
}

// putResource 创建或更新 APISIX 资源 (upstream/route)
func (a *ApisixGateway) putResource(ctx context.Context, resource, id string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/apisix/admin/%s/%s", a.config.AdminURL, resource, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if a.config.APIKey != "" {
		req.Header.Set("X-API-KEY", a.config.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apisix %s %s/%s returned %d: %s", http.MethodPut, resource, id, resp.StatusCode, string(respBody))
	}
	return nil
}

// deleteResource 删除 APISIX 资源
func (a *ApisixGateway) deleteResource(ctx context.Context, resource, id string) error {
	url := fmt.Sprintf("%s/apisix/admin/%s/%s", a.config.AdminURL, resource, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if a.config.APIKey != "" {
		req.Header.Set("X-API-KEY", a.config.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apisix DELETE %s/%s returned %d: %s", resource, id, resp.StatusCode, string(respBody))
	}
	return nil
}
