package breaker

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultK            = 1.5
	defaultBuckets      = 40
	defaultBucketTimeMs = 250
	defaultMinReqs      = 100
)

// GoogleBreakerConfig 熔断器配置。
type GoogleBreakerConfig struct {
	K            float64
	Buckets      int
	BucketTimeMs int
	MinReqs      int64
}

// DefaultGoogleBreakerConfig 返回默认熔断器配置。
func DefaultGoogleBreakerConfig() GoogleBreakerConfig {
	return GoogleBreakerConfig{
		K:            defaultK,
		Buckets:      defaultBuckets,
		BucketTimeMs: defaultBucketTimeMs,
		MinReqs:      defaultMinReqs,
	}
}

// rollingWindow 滑动窗口
type rollingWindow struct {
	mu         sync.Mutex
	buckets    []bucket
	size       int
	bucketTime time.Duration
	offset     int
	lastTime   time.Time
}

type bucket struct {
	accepts int64
	total   int64
}

func newRollingWindow(size int, bucketTime time.Duration) *rollingWindow {
	return &rollingWindow{
		buckets:    make([]bucket, size),
		size:       size,
		bucketTime: bucketTime,
		lastTime:   time.Now(),
	}
}

func (rw *rollingWindow) advance() {
	now := time.Now()
	elapsed := int(now.Sub(rw.lastTime) / rw.bucketTime)
	if elapsed <= 0 {
		return
	}
	if elapsed > rw.size {
		elapsed = rw.size
	}
	for i := 0; i < elapsed; i++ {
		rw.offset = (rw.offset + 1) % rw.size
		rw.buckets[rw.offset] = bucket{}
	}
	rw.lastTime = now
}

func (rw *rollingWindow) add(accepts, total int64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.advance()
	rw.buckets[rw.offset].accepts += accepts
	rw.buckets[rw.offset].total += total
}

func (rw *rollingWindow) sum() (accepts, total int64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.advance()
	for _, b := range rw.buckets {
		accepts += b.accepts
		total += b.total
	}
	return
}

// GoogleBreaker Google SRE 风格熔断器
// 丢弃概率: max(0, (requests - K * accepts) / (requests + 1))
type GoogleBreaker struct {
	window  *rollingWindow
	k       float64
	minReqs int64
}

func NewGoogleBreaker() Breaker {
	return NewGoogleBreakerWithConfig(DefaultGoogleBreakerConfig())
}

// NewGoogleBreakerWithConfig 使用指定配置创建熔断器。
func NewGoogleBreakerWithConfig(config GoogleBreakerConfig) Breaker {
	config = normalizeGoogleBreakerConfig(config)

	return &GoogleBreaker{
		window:  newRollingWindow(config.Buckets, time.Duration(config.BucketTimeMs)*time.Millisecond),
		k:       config.K,
		minReqs: config.MinReqs,
	}
}

func normalizeGoogleBreakerConfig(config GoogleBreakerConfig) GoogleBreakerConfig {
	defaults := DefaultGoogleBreakerConfig()
	if config.K <= 0 {
		config.K = defaults.K
	}
	if config.Buckets <= 0 {
		config.Buckets = defaults.Buckets
	}
	if config.BucketTimeMs <= 0 {
		config.BucketTimeMs = defaults.BucketTimeMs
	}
	if config.MinReqs <= 0 {
		config.MinReqs = defaults.MinReqs
	}

	return config
}

func (gb *GoogleBreaker) Allow() (Promise, error) {
	if err := gb.accept(); err != nil {
		return nil, err
	}
	return &googlePromise{breaker: gb}, nil
}

func (gb *GoogleBreaker) Do(req func() error) error {
	return gb.DoWithAcceptable(req, func(err error) bool { return false })
}

func (gb *GoogleBreaker) DoWithAcceptable(req func() error, acceptable func(err error) bool) error {
	promise, err := gb.Allow()
	if err != nil {
		return err
	}

	err = req()
	if err == nil || acceptable(err) {
		promise.Accept()
	} else {
		promise.Reject(err)
	}
	return err
}

func (gb *GoogleBreaker) accept() error {
	accepts, total := gb.window.sum()
	// 请求量不足，不熔断
	if total < gb.minReqs {
		return nil
	}

	dropRatio := gb.dropRatio(accepts, total)
	if dropRatio <= 0 {
		return nil
	}

	shouldDrop := rand.Float64() < dropRatio

	if shouldDrop {
		return ErrServiceUnavailable
	}
	return nil
}

func (gb *GoogleBreaker) dropRatio(accepts, total int64) float64 {
	ratio := math.Max(0, (float64(total)-gb.k*float64(accepts))/float64(total+1))
	return ratio
}

func (gb *GoogleBreaker) markSuccess() {
	gb.window.add(1, 1)
}

func (gb *GoogleBreaker) markFailure() {
	gb.window.add(0, 1)
}

// googlePromise 请求承诺实现
type googlePromise struct {
	breaker *GoogleBreaker
}

func (p *googlePromise) Accept() {
	p.breaker.markSuccess()
}

func (p *googlePromise) Reject(_ error) {
	p.breaker.markFailure()
}
