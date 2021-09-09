package utils

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &RateLimiter{}

// RateLimiter scaffolds the default rate limiter.
type RateLimiter struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *RateLimiter) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "utils", "rate_limiter.go")

	f.TemplateBody = controllerRateLimiterTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const controllerRateLimiterTemplate = `{{ .Boilerplate }}

package utils

import (
	"math"
	"sync"
	"time"
)

type DefaultRateLimiter struct {
	requeuesLock sync.Mutex
	requeues     map[interface{}]int
	modifier     map[interface{}]int

	baseDelay time.Duration
	maxDelay  time.Duration
}

func NewDefaultRateLimiter(baseDelay, maxDelay time.Duration) *DefaultRateLimiter {
	return &DefaultRateLimiter{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		requeues: map[interface{}]int{},
		modifier: map[interface{}]int{},
	}
}

func (r *DefaultRateLimiter) When(item interface{}) time.Duration {
	r.requeuesLock.Lock()
	defer r.requeuesLock.Unlock()

	exp := r.modifier[item]
	r.requeues[item]++

	if r.requeues[item]%16 == 0 {
		r.modifier[item]++
	}

	// The backoff is capped such that 'calculated' value never overflows.
	backoff := float64(r.baseDelay.Nanoseconds()) * math.Pow(2, float64(exp))
	if backoff > math.MaxInt64 {
		return r.maxDelay
	}

	calculated := time.Duration(backoff)
	if calculated > r.maxDelay {
		return r.maxDelay
	}

	return calculated
}

func (r *DefaultRateLimiter) NumRequeues(item interface{}) int {
	r.requeuesLock.Lock()
	defer r.requeuesLock.Unlock()

	return r.requeues[item]
}

func (r *DefaultRateLimiter) Forget(item interface{}) {
	r.requeuesLock.Lock()
	defer r.requeuesLock.Unlock()

	delete(r.requeues, item)
}
`
