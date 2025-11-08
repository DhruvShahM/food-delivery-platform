package public

import (
	"context"
	"database/sql"
	"food-delivery-platform/services/auth-service/internal/handler"
	"food-delivery-platform/services/auth-service/internal/proto"
	"food-delivery-platform/services/auth-service/internal/repository"
	"go.uber.org/zap"
)

// AuthService provides public access to auth functionality for integration testing
type AuthService interface {
	Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error)
	Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error)
}

// NewAuthService creates a new auth service instance
func NewAuthService(db *sql.DB, logger *zap.Logger) AuthService {
	repo := repository.NewUserRepo(db, logger)
	handler := handler.NewAuthHandler(repo, "secret", logger)
	return &authServiceImpl{handler: handler}
}

type authServiceImpl struct {
	handler *handler.AuthHandler
}

func (a *authServiceImpl) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	return a.handler.Register(ctx, req)
}

func (a *authServiceImpl) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	return a.handler.Login(ctx, req)
}
