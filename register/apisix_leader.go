package register

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type apisixLeaderController struct {
	gateway *ApisixGateway
	service *GatewayService

	leaseName      string
	leaseNamespace string
	identity       string

	leaseDuration     time.Duration
	renewDeadline     time.Duration
	retryPeriod       time.Duration
	reconcileInterval time.Duration

	mu       sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	started  bool
	isLeader atomic.Bool
}

func newApisixLeaderController(gateway *ApisixGateway, service *GatewayService) (*apisixLeaderController, error) {
	namespace := strings.TrimSpace(gateway.config.LeaderLeaseNamespace)
	if namespace == "" {
		namespace = defaultLeaseNamespace()
	}
	if namespace == "" {
		return nil, fmt.Errorf("leader mode requires lease namespace (set leader_lease_namespace or POD_NAMESPACE)")
	}

	leaseName := strings.TrimSpace(gateway.config.LeaderLeaseName)
	if leaseName == "" {
		leaseName = fmt.Sprintf("%s-apisix-lock", service.ID)
	}

	identity := strings.TrimSpace(gateway.config.LeaderIdentity)
	if identity == "" {
		identity = defaultLeaderIdentity()
	}
	if identity == "" {
		return nil, fmt.Errorf("leader mode requires identity (set leader_identity or POD_NAME)")
	}

	controller := &apisixLeaderController{
		gateway:           gateway,
		service:           service,
		leaseName:         leaseName,
		leaseNamespace:    namespace,
		identity:          identity,
		leaseDuration:     durationOrDefault(gateway.config.LeaseDurationSeconds, 15),
		renewDeadline:     durationOrDefault(gateway.config.RenewDeadlineSeconds, 10),
		retryPeriod:       durationOrDefault(gateway.config.RetryPeriodSeconds, 2),
		reconcileInterval: durationOrDefault(gateway.config.ReconcileIntervalSeconds, 30),
		done:              make(chan struct{}),
	}
	return controller, nil
}

func (c *apisixLeaderController) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return nil
	}

	restCfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("leader mode requires in-cluster kubernetes config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes client failed: %w", err)
	}

	runCtx := context.Background()
	if ctx != nil {
		runCtx = ctx
	}
	runCtx, c.cancel = context.WithCancel(runCtx)
	c.started = true

	go c.run(runCtx, clientset)
	return nil
}

func (c *apisixLeaderController) run(ctx context.Context, clientset *kubernetes.Clientset) {
	defer close(c.done)
	defer func() {
		if recover() != nil {
			c.isLeader.Store(false)
		}
	}()

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      c.leaseName,
			Namespace: c.leaseNamespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: c.identity,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		LeaseDuration:   c.leaseDuration,
		RenewDeadline:   c.renewDeadline,
		RetryPeriod:     c.retryPeriod,
		ReleaseOnCancel: true,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leadCtx context.Context) {
				c.isLeader.Store(true)
				_ = c.gateway.registerServiceDirect(leadCtx, c.service)

				ticker := time.NewTicker(c.reconcileInterval)
				defer ticker.Stop()

				for {
					select {
					case <-leadCtx.Done():
						return
					case <-ticker.C:
						_ = c.gateway.registerServiceDirect(leadCtx, c.service)
					}
				}
			},
			OnStoppedLeading: func() {
				c.isLeader.Store(false)
			},
		},
	})
}

func (c *apisixLeaderController) Stop() {
	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return
	}
	cancel := c.cancel
	done := c.done
	c.started = false
	c.cancel = nil
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	<-done
}

func (c *apisixLeaderController) IsLeader() bool {
	return c.isLeader.Load()
}

func durationOrDefault(seconds int, fallbackSeconds int) time.Duration {
	if seconds <= 0 {
		seconds = fallbackSeconds
	}
	return time.Duration(seconds) * time.Second
}

func defaultLeaseNamespace() string {
	if v := strings.TrimSpace(os.Getenv("POD_NAMESPACE")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("NAMESPACE")); v != "" {
		return v
	}
	bytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(bytes))
}

func defaultLeaderIdentity() string {
	if v := strings.TrimSpace(os.Getenv("POD_NAME")); v != "" {
		return v
	}
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(hostname)
}
