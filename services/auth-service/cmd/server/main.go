package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"food-delivery-platform/services/auth-service/internal/config"
	"food-delivery-platform/services/auth-service/internal/handler"
	"food-delivery-platform/services/auth-service/internal/proto"
	"food-delivery-platform/services/auth-service/internal/repository"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"
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

var tracer = otel.Tracer("auth-service")

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
		resource.WithAttributes(semconv.ServiceNameKey.String("auth-service")),
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
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	defer cfg.Logger.Sync()

	logger := cfg.Logger

	// Initialize OTEL
	tpShutdown := initTracer(logger)
	defer tpShutdown()

	// Initialize DB
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		logger.Error("Failed to open DB connection", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping DB", zap.Error(err))
		os.Exit(1)
	}

	repo := repository.NewUserRepo(db, logger)
	if err := repo.Init(); err != nil {
		logger.Error("Failed to initialize DB tables", zap.Error(err))
		os.Exit(1)
	}

	// Initialize handler
	h := handler.NewAuthHandler(repo, cfg.JWTSecret, logger)

	// gRPC server
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("Failed to listen", zap.Error(err))
		os.Exit(1)
	}

	// gRPC server setup
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1<<20), // 1MB
		grpc.MaxSendMsgSize(1<<20),
	)

	// Register service
	proto.RegisterAuthServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		logger.Info("gRPC server starting", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// HTTP server for metrics
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics endpoint\n"))
	})
	httpServer := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("HTTP server starting", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown gRPC
	grpcServer.GracefulStop()

	// Shutdown HTTP
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Shutdown complete")
}
