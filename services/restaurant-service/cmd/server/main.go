package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"food-delivery-platform/services/restaurant-service/internal/config"
	"food-delivery-platform/services/restaurant-service/internal/handler"
	"food-delivery-platform/services/restaurant-service/internal/proto"
	"food-delivery-platform/services/restaurant-service/internal/repository"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var tracer = otel.Tracer("restaurant-service")

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
	resource, _ := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String("restaurant-service")),
	)
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return func() { tp.Shutdown(context.Background()) }
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error", err)
	}
	defer cfg.Logger.Sync()

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

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		cfg.Logger.Fatal("Redis ping", zap.Error(err))
	}

	repo := repository.NewMenuRepo(db, cfg.Logger)
	if err := repo.Init(); err != nil {
		cfg.Logger.Fatal("Repo init", zap.Error(err))
	}

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		cfg.Logger.Fatal("Listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
		),
	)
	proto.RegisterRestaurantServiceServer(grpcServer, handler.NewRestaurantHandler(repo, redisClient, cfg.Logger))
	grpc_health_v1.RegisterHealthServer(grpcServer, &healthServer{})
	reflection.Register(grpcServer)

	cfg.Logger.Info("gRPC server starting", zap.String("port", cfg.GRPCPort))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			cfg.Logger.Fatal("gRPC serve", zap.Error(err))
		}
	}()

	r := gin.Default()
	r.Use(handler.CORS())
	r.GET("/metrics", func(c *gin.Context) {
		_, span := tracer.Start(c.Request.Context(), "metrics")
		defer span.End()
		span.SetAttributes(attribute.String("method", c.Request.Method))
		c.Status(http.StatusOK)
	})

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
