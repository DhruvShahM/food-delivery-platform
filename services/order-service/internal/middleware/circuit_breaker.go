// services/order-service/internal/middleware/circuit_breaker.go
package middleware

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CircuitBreakerConfig struct {
	Name          string        // Circuit breaker name
	MaxRequests   uint32        // Max concurrent requests in half-open state
	Interval      time.Duration // Reset interval
	Timeout       time.Duration // Timeout for requests
	FailureRatio  float64       // Failure threshold (0.0-1.0)
	Logger        *zap.Logger
}

type CircuitBreaker struct {
	cb     *gobreaker.CircuitBreaker
	config *CircuitBreakerConfig
	logger *zap.Logger
}

func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= config.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			config.Logger.Info("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	return &CircuitBreaker{
		cb:     gobreaker.NewCircuitBreaker(settings),
		config: config,
		logger: config.Logger,
	}
}

// GRPCCircuitBreakerInterceptor creates a gRPC interceptor with circuit breaker
func (cb *CircuitBreaker) GRPCCircuitBreakerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Execute request through circuit breaker
		result, err := cb.cb.Execute(func() (interface{}, error) {
			return handler(ctx, req)
		})

		if err != nil {
			// Check if it's a circuit breaker error
			if err == gobreaker.ErrOpenState {
				cb.logger.Warn("Circuit breaker is open, rejecting request",
					zap.String("method", info.FullMethod),
					zap.String("breaker", cb.config.Name),
				)
				return nil, status.Error(codes.Unavailable, "Service temporarily unavailable")
			}
			if err == gobreaker.ErrTooManyRequests {
				cb.logger.Warn("Circuit breaker too many requests",
					zap.String("method", info.FullMethod),
					zap.String("breaker", cb.config.Name),
				)
				return nil, status.Error(codes.ResourceExhausted, "Too many requests")
			}
			// Return the original error
			return nil, err
		}

		return result, nil
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() gobreaker.State {
	return cb.cb.State()
}

// GetCounts returns the current circuit breaker counts
func (cb *CircuitBreaker) GetCounts() gobreaker.Counts {
	return cb.cb.Counts()
}