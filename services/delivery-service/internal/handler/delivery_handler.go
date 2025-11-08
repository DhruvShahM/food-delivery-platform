package handler

import (
	"context"
	"fmt"
	"math/rand"

	"food-delivery-platform/services/delivery-service/internal/proto"
	"food-delivery-platform/services/delivery-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeliveryHandler struct {
	proto.UnimplementedDeliveryServiceServer
	repo   *repository.DeliveryRepo
	logger *zap.Logger
}

func NewDeliveryHandler(repo *repository.DeliveryRepo, logger *zap.Logger) *DeliveryHandler {
	return &DeliveryHandler{repo: repo, logger: logger}
}

func (h *DeliveryHandler) AssignDelivery(ctx context.Context, req *proto.AssignDeliveryRequest) (*proto.AssignDeliveryResponse, error) {
	agentId := fmt.Sprintf("agent_%d", rand.Int31())
	err := h.repo.Assign(req.OrderId, agentId)
	if err != nil {
		h.logger.Error("Assign error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}
	h.logger.Info("Delivery assigned", zap.String("order_id", req.OrderId), zap.String("agent_id", agentId))
	return &proto.AssignDeliveryResponse{AgentId: agentId}, nil
}

func (h *DeliveryHandler) UpdateStatus(ctx context.Context, req *proto.UpdateStatusRequest) (*proto.UpdateStatusResponse, error) {
	h.logger.Info("Update delivery status", zap.String("delivery_id", req.DeliveryId), zap.String("status", req.Status))

	err := h.repo.UpdateDeliveryStatus(req.DeliveryId, req.Status)
	if err != nil {
		h.logger.Error("Update delivery status error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	h.logger.Info("Delivery status updated", zap.String("delivery_id", req.DeliveryId), zap.String("status", req.Status))
	return &proto.UpdateStatusResponse{Success: true}, nil
}

func (h *DeliveryHandler) GetDeliveries(ctx context.Context, req *proto.GetDeliveriesRequest) (*proto.GetDeliveriesResponse, error) {
	h.logger.Info("Get deliveries", zap.String("agent_id", req.AgentId))

	deliveries, err := h.repo.GetDeliveriesByAgentID(req.AgentId)
	if err != nil {
		h.logger.Error("Get deliveries error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	h.logger.Info("Deliveries retrieved", zap.String("agent_id", req.AgentId), zap.Int("count", len(deliveries)))
	return &proto.GetDeliveriesResponse{Deliveries: deliveries}, nil
}
