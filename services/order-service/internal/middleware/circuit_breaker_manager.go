// services/order-service/internal/middleware/circuit_breaker_manager.go
package middleware

import (
	"sync"

	"go.uber.org/zap"
)

type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	logger   *zap.Logger
	mu       sync.RWMutex
}

func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreateBreaker gets an existing breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreateBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}

	config.Name = name
	breaker := NewCircuitBreaker(config)
	cbm.breakers[name] = breaker

	cbm.logger.Info("Created new circuit breaker",
		zap.String("name", name),
		zap.Duration("interval", config.Interval),
		zap.Float64("failure_ratio", config.FailureRatio),
	)

	return breaker
}

// GetBreaker returns an existing circuit breaker
func (cbm *CircuitBreakerManager) GetBreaker(name string) (*CircuitBreaker, bool) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	breaker, exists := cbm.breakers[name]
	return breaker, exists
}

// GetAllBreakers returns all circuit breakers with their states
func (cbm *CircuitBreakerManager) GetAllBreakers() map[string]map[string]interface{} {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for name, breaker := range cbm.breakers {
		counts := breaker.GetCounts()
		result[name] = map[string]interface{}{
			"state":         breaker.GetState().String(),
			"requests":      counts.Requests,
			"total_success": counts.TotalSuccesses,
			"total_failure": counts.TotalFailures,
			"consecutive_success": counts.ConsecutiveSuccesses,
			"consecutive_failure": counts.ConsecutiveFailures,
		}
	}
	return result
}