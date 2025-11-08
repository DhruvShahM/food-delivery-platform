package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"food-delivery-platform/services/order-service/internal/config"
	"food-delivery-platform/services/order-service/internal/handler"
	"food-delivery-platform/services/order-service/internal/middleware"
	"food-delivery-platform/services/order-service/internal/proto"
	"food-delivery-platform/services/order-service/internal/repository"

	"food-delivery-platform/common/pkg/kafka"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	// "github.com/sony/gobreaker"
)

var tracer = otel.Tracer("order-service")

type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func initTracer(logger *zap.Logger) func() {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		logger.Fatal("Jaeger exporter error", zap.Error(err))
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Tracer shutdown error", zap.Error(err))
		}
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error", err)
	}

	tpShutdown := initTracer(cfg.Logger)
	defer tpShutdown()

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		cfg.Logger.Fatal("DB connect", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		cfg.Logger.Fatal("DB ping", zap.Error(err))
	}

	producer := kafka.NewWriter(cfg.KafkaBrokers, cfg.KafkaTopic)
	repo := repository.NewOrderRepo(db, producer, cfg.Logger)
	if err := repo.Init(); err != nil {
		cfg.Logger.Fatal("Repo init", zap.Error(err))
	}

	// Initialize Redis client for middleware
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// Initialize auth service
	authConfig := &middleware.AuthConfig{
		JWTSecret:   cfg.JWTSecret,
		RedisClient: redisClient,
		TokenExpiry: 24 * time.Hour,
		Logger:      cfg.Logger,
	}
	authService := middleware.NewAuthService(authConfig)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.Logger)

	// Initialize circuit breaker manager
	cbManager := middleware.NewCircuitBreakerManager(cfg.Logger)

	// Initialize HTTP handler
	orderHandler := handler.NewOrderHandler(repo, cfg.Logger)

	// Create circuit breaker configs for different services
	authCB := cbManager.GetOrCreateBreaker("auth-service", &middleware.CircuitBreakerConfig{
		MaxRequests:  3,                // Max requests in half-open state
		Interval:     60 * time.Second, // Reset every 60 seconds
		Timeout:      30 * time.Second, // Request timeout
		FailureRatio: 0.5,              // Open circuit if 50% failures
		Logger:       cfg.Logger,
	})

	// paymentCB := cbManager.GetOrCreateBreaker("payment-service", &middleware.CircuitBreakerConfig{
	// 	MaxRequests:  5,
	// 	Interval:     30 * time.Second,
	// 	Timeout:      10 * time.Second,
	// 	FailureRatio: 0.3, // More sensitive for payment
	// 	Logger:       cfg.Logger,
	// })

	// deliveryCB := cbManager.GetOrCreateBreaker("delivery-service", &middleware.CircuitBreakerConfig{
	// 	MaxRequests:  3,
	// 	Interval:     45 * time.Second,
	// 	Timeout:      20 * time.Second,
	// 	FailureRatio: 0.4,
	// 	Logger:       cfg.Logger,
	// })

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		cfg.Logger.Fatal("Listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			rateLimiter.GRPCIPRateLimit(1000, time.Minute), // 1000 requests per minute per IP
			authService.GRPCAuthInterceptor(),               // Require authentication for all gRPC calls
			rateLimiter.GRPCUserRateLimit(5000, time.Hour),  // 5000 requests per hour per user
			// Add circuit breaker interceptor (applied to all methods)
			authCB.GRPCCircuitBreakerInterceptor(),
		),
	)
	proto.RegisterOrderServiceServer(grpcServer, orderHandler)
	grpc_health_v1.RegisterHealthServer(grpcServer, &healthServer{})
	reflection.Register(grpcServer)

	cfg.Logger.Info("gRPC server starting", zap.String("port", cfg.GRPCPort))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			cfg.Logger.Fatal("gRPC serve", zap.Error(err))
		}
	}()

	// Setup Gin with middleware
	r := gin.Default()
	r.Use(handler.CORS())
	
	// Apply global rate limiting (100 requests per minute per IP)
	r.Use(rateLimiter.IPRateLimit(100, time.Minute))

	// Public endpoints (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.GET("/metrics", func(c *gin.Context) {
		_, span := tracer.Start(c.Request.Context(), "metrics")
		defer span.End()
		span.SetAttributes(attribute.String("method", c.Request.Method))
		c.Status(http.StatusOK)
	})

	// Auth endpoints
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", handler.LoginHandler(authService))
		authGroup.POST("/register", handler.RegisterHandler(authService))
		authGroup.POST("/logout", authService.LogoutHandler())
	}

	// Circuit breaker status endpoint
	r.GET("/circuit-breakers", func(c *gin.Context) {
		status := cbManager.GetAllBreakers()
		c.JSON(http.StatusOK, gin.H{
			"circuit_breakers": status,
			"timestamp":        time.Now(),
		})
	})

	// Protected API endpoints
	api := r.Group("/api/v1")
	api.Use(authService.JWTMiddleware()) // Require authentication
	api.Use(rateLimiter.UserRateLimit(1000, time.Hour)) // 1000 requests per hour per user
	{
		// Apply stricter rate limiting to sensitive endpoints
		orders := api.Group("/orders")
		orders.Use(rateLimiter.EndpointRateLimit(50, time.Minute)) // 50 requests per minute per endpoint
		{
			orders.POST("", orderHandler.PlaceOrderHTTP)
			orders.GET("", orderHandler.GetOrdersHTTP)
			orders.PUT("/:id/status", authService.RoleBasedAuthMiddleware("admin", "restaurant"), orderHandler.UpdateOrderStatusHTTP)
		}
	}

	srv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	cfg.Logger.Info("HTTP server starting", zap.String("port", cfg.HTTPPort))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		cfg.Logger.Fatal("HTTP serve", zap.Error(err))
	}
}