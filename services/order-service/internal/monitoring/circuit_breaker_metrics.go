// services/order-service/internal/monitoring/circuit_breaker_metrics.go
package monitoring

import (
	"net/http"

	"food-delivery-platform/services/order-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of circuit breaker (0=closed, 1=open, 2=half-open)",
		},
		[]string{"name"},
	)

	circuitBreakerRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_requests_total",
			Help: "Total number of requests through circuit breaker",
		},
		[]string{"name", "state"},
	)

	circuitBreakerFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures_total",
			Help: "Total number of failed requests through circuit breaker",
		},
		[]string{"name"},
	)
)

func init() {
	prometheus.MustRegister(circuitBreakerState, circuitBreakerRequests, circuitBreakerFailures)
}

// UpdateCircuitBreakerMetrics updates Prometheus metrics for circuit breakers
func UpdateCircuitBreakerMetrics(cbManager *middleware.CircuitBreakerManager) {
	breakers := cbManager.GetAllBreakers()

	for name, info := range breakers {
		state := 0.0
		switch info["state"].(string) {
		case "closed":
			state = 0.0
		case "open":
			state = 1.0
		case "half-open":
			state = 2.0
		}

		circuitBreakerState.WithLabelValues(name).Set(state)
		circuitBreakerRequests.WithLabelValues(name, info["state"].(string)).Add(float64(info["requests"].(uint32)))
		circuitBreakerFailures.WithLabelValues(name).Add(float64(info["total_failure"].(uint32)))
	}
}

// SetupMetricsServer sets up a metrics endpoint
func SetupMetricsServer(cbManager *middleware.CircuitBreakerManager) *gin.Engine {
	r := gin.New()

	// Update metrics every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UpdateCircuitBreakerMetrics(cbManager)
		}
	}()

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return r
}
