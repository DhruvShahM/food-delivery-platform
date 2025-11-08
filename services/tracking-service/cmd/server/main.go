package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"food-delivery-platform/services/tracking-service/internal/config"
	"food-delivery-platform/services/tracking-service/internal/handler"
	"food-delivery-platform/services/tracking-service/internal/proto"
	"food-delivery-platform/services/tracking-service/internal/repository"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var tracer = zap.NewNop().Sugar() // Simple placeholder

type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func initTracer(logger *zap.Logger) func() {
	// Simple placeholder - no complex tracing
	return func() {}
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

	repo := repository.NewTrackingRepo(db, cfg.Logger)
	if err := repo.Init(); err != nil {
		cfg.Logger.Fatal("Repo init", zap.Error(err))
	}

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		cfg.Logger.Fatal("Listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	proto.RegisterTrackingServiceServer(grpcServer, handler.NewTrackingHandler(repo, cfg.Logger))
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
