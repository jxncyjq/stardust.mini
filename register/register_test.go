package register

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jxncyjq/stardust.mini/logs"
)

func TestMain(m *testing.M) {
	logs.Init([]byte(`{"level":-1}`))
	os.Exit(m.Run())
}

// MockRegister 模拟注册器用于测试
type MockRegister struct {
	mu       sync.RWMutex
	services map[string][]*ServiceInfo
	watchers map[string][]chan []*ServiceInfo
}

func NewMockRegister() *MockRegister {
	return &MockRegister{
		services: make(map[string][]*ServiceInfo),
		watchers: make(map[string][]chan []*ServiceInfo),
	}
}

func (m *MockRegister) Register(ctx context.Context, info *ServiceInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services[info.Name] = append(m.services[info.Name], info)
	m.notifyWatchers(info.Name)
	return nil
}

func (m *MockRegister) Deregister(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, services := range m.services {
		for i, s := range services {
			if s.ID == serviceID {
				m.services[name] = append(services[:i], services[i+1:]...)
				m.notifyWatchers(name)
				return nil
			}
		}
	}
	return nil
}

func (m *MockRegister) GetService(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services[serviceName], nil
}

func (m *MockRegister) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch := make(chan []*ServiceInfo, 10)
	m.watchers[serviceName] = append(m.watchers[serviceName], ch)
	if services, ok := m.services[serviceName]; ok {
		ch <- services
	}
	return ch, nil
}

func (m *MockRegister) notifyWatchers(serviceName string) {
	if watchers, ok := m.watchers[serviceName]; ok {
		services := m.services[serviceName]
		for _, ch := range watchers {
			select {
			case ch <- services:
			default:
			}
		}
	}
}

func (m *MockRegister) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, watchers := range m.watchers {
		for _, ch := range watchers {
			close(ch)
		}
	}
	return nil
}

func TestServiceInfo(t *testing.T) {
	info := &ServiceInfo{
		Name:    "test-service",
		ID:      "test-1",
		Address: "127.0.0.1",
		Port:    8080,
		Tags:    []string{"http", "api"},
		Meta:    map[string]string{"version": "1.0"},
	}

	if info.Name != "test-service" {
		t.Errorf("Expected name 'test-service', got '%s'", info.Name)
	}
	if info.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", info.Port)
	}
}

func TestMockRegister(t *testing.T) {
	ctx := context.Background()
	reg := NewMockRegister()
	defer reg.Close()

	// 测试注册
	info := &ServiceInfo{
		Name:    "test-service",
		ID:      "test-1",
		Address: "127.0.0.1",
		Port:    8080,
	}

	err := reg.Register(ctx, info)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试获取服务
	services, err := reg.GetService(ctx, "test-service")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	if services[0].ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", services[0].ID)
	}

	// 测试注销
	err = reg.Deregister(ctx, "test-1")
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	services, _ = reg.GetService(ctx, "test-service")
	if len(services) != 0 {
		t.Errorf("Expected 0 services after deregister, got %d", len(services))
	}
}

func TestServiceRegistry(t *testing.T) {
	reg := NewMockRegister()
	defer reg.Close()

	registry := NewServiceRegistry(reg)

	// 测试注册服务
	err := registry.Register("my-service", "127.0.0.1", 8080, []string{"http"}, nil)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试发现服务
	services, err := registry.Discover("my-service")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	// 测试注销
	err = registry.Deregister()
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	services, _ = registry.Discover("my-service")
	if len(services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(services))
	}
}

func TestWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reg := NewMockRegister()
	defer reg.Close()

	// 开始监听
	ch, err := reg.Watch(ctx, "watch-service")
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// 注册服务
	info := &ServiceInfo{
		Name:    "watch-service",
		ID:      "watch-1",
		Address: "127.0.0.1",
		Port:    8080,
	}
	reg.Register(ctx, info)

	// 接收通知
	services := <-ch
	if len(services) != 1 {
		t.Errorf("Expected 1 service in watch, got %d", len(services))
	}
}
