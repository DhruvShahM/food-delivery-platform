package handler

import (
	"context"
	"fmt"
	"math/rand"

	"food-delivery-platform/services/restaurant-service/internal/proto"
	"food-delivery-platform/services/restaurant-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RestaurantHandler struct {
	proto.UnimplementedRestaurantServiceServer
	repo   *repository.MenuRepo
	redis  *redis.Client
	logger *zap.Logger
}

func NewRestaurantHandler(repo *repository.MenuRepo, redis *redis.Client, logger *zap.Logger) *RestaurantHandler {
	return &RestaurantHandler{repo: repo, redis: redis, logger: logger}
}

func (h *RestaurantHandler) GetMenu(ctx context.Context, req *proto.GetMenuRequest) (*proto.GetMenuResponse, error) {
	items, err := h.repo.GetMenu(req.RestaurantId)
	if err != nil {
		h.logger.Error("Get menu error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return &proto.GetMenuResponse{Items: items}, nil
}

func (h *RestaurantHandler) UpdateAvailability(ctx context.Context, req *proto.UpdateAvailabilityRequest) (*proto.UpdateAvailabilityResponse, error) {
	err := h.repo.UpdateAvailability(req.ItemId, req.Available)
	if err != nil {
		h.logger.Error("Update availability error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return &proto.UpdateAvailabilityResponse{Success: true}, nil
}

func (h *RestaurantHandler) CreateMenuItem(ctx context.Context, req *proto.CreateMenuItemRequest) (*proto.CreateMenuItemResponse, error) {
	itemId := fmt.Sprintf("item_%d", rand.Int63())
	err := h.repo.CreateMenuItem(req.RestaurantId, itemId, req.Name, req.Price)
	if err != nil {
		h.logger.Error("Create item error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return &proto.CreateMenuItemResponse{ItemId: itemId}, nil
}
