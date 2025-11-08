package handler

import (
	"context"
	"time"

	"food-delivery-platform/services/auth-service/internal/proto"
	"food-delivery-platform/services/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthHandler struct {
	proto.UnimplementedAuthServiceServer
	repo    *repository.UserRepo
	secret  []byte
	logger  *zap.Logger
}

func NewAuthHandler(repo *repository.UserRepo, secret string, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		repo:    repo,
		secret:  []byte(secret),
		logger:  logger,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	h.logger.Info("Register attempt", zap.String("email", req.Email))

	// Check if user already exists
	existing, err := h.repo.GetUserByEmail(req.Email)
	if err == nil && existing != nil {
		return &proto.RegisterResponse{Error: "User already exists"}, nil
	}

	// Create new user
	err = h.repo.CreateUser(req.Email, req.Password)
	if err != nil {
		h.logger.Error("Register failed", zap.String("email", req.Email), zap.Error(err))
		return &proto.RegisterResponse{Error: "Registration failed"}, nil
	}

	// Auto-login after registration
	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		h.logger.Error("Auto-login failed", zap.String("email", req.Email), zap.Error(err))
		return &proto.RegisterResponse{Error: "Registration successful, please login"}, nil
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.secret)
	if err != nil {
		h.logger.Error("JWT error", zap.Error(err))
		return &proto.RegisterResponse{Error: "Registration successful, please login"}, nil
	}

	h.logger.Info("Register success", zap.Int("user_id", user.ID))
	return &proto.RegisterResponse{Token: tokenStr}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	h.logger.Info("Login attempt", zap.String("email", req.Email))

	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("email", req.Email), zap.Error(err))
		return &proto.LoginResponse{Error: "Invalid credentials"}, nil
	}

	if user.Password != req.Password {
		h.logger.Warn("Password mismatch", zap.String("email", req.Email))
		return &proto.LoginResponse{Error: "Invalid credentials"}, nil
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.secret)
	if err != nil {
		h.logger.Error("JWT error", zap.Error(err))
		return &proto.LoginResponse{Error: "Internal error"}, nil
	}

	h.logger.Info("Login success", zap.Int("user_id", user.ID))
	return &proto.LoginResponse{Token: tokenStr}, nil
}