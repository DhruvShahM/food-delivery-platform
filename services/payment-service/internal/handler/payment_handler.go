package handler

import (
	"context"
	"fmt"
	"food-delivery-platform/services/payment-service/internal/proto"
	"food-delivery-platform/services/payment-service/internal/repository"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	proto.UnimplementedPaymentServiceServer
	repo   *repository.PaymentRepo
	redis  *redis.Client
	logger *zap.Logger
}

func NewPaymentHandler(repo *repository.PaymentRepo, redis *redis.Client, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{repo: repo, redis: redis, logger: logger}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *proto.ProcessPaymentRequest) (*proto.ProcessPaymentResponse, error) {
	// For demo purposes, assume user_id is part of order_id or extract from context
	// In real implementation, you'd get user_id from authentication context
	userId := "user123" // TODO: Get from auth context

	balance, err := h.redis.Get(ctx, "wallet:"+userId).Float64()
	if err != nil || balance < req.Amount {
		h.logger.Warn("Insufficient balance", zap.String("order_id", req.OrderId), zap.Float64("amount", req.Amount))
		return &proto.ProcessPaymentResponse{Status: "failed", Error: "Insufficient balance"}, nil
	}

	err = h.repo.Process(userId, req.Amount)
	if err != nil {
		h.logger.Error("Process error", zap.Error(err))
		return &proto.ProcessPaymentResponse{Status: "failed", Error: "Internal error"}, nil
	}

	// Update balance
	_, err = h.redis.DecrBy(ctx, "wallet:"+userId, int64(req.Amount)).Result()
	if err != nil {
		h.logger.Error("Redis update error", zap.Error(err))
	}

	paymentId := fmt.Sprintf("payment_%d", time.Now().UnixNano())
	h.logger.Info("Payment processed", zap.String("payment_id", paymentId), zap.String("order_id", req.OrderId), zap.Float64("amount", req.Amount))
	return &proto.ProcessPaymentResponse{PaymentId: paymentId, Status: "completed"}, nil
}

func (h *PaymentHandler) GetBalance(ctx context.Context, req *proto.GetBalanceRequest) (*proto.GetBalanceResponse, error) {
	balance, err := h.redis.Get(ctx, "wallet:"+req.UserId).Float64()
	if err != nil {
		// If not in Redis, get from database
		balance, err = h.repo.GetBalance(req.UserId)
		if err != nil {
			h.logger.Error("Get balance error", zap.Error(err))
			return &proto.GetBalanceResponse{Error: "Internal error"}, nil
		}
		// Cache in Redis
		h.redis.Set(ctx, "wallet:"+req.UserId, balance, 0)
	}

	h.logger.Info("Balance retrieved", zap.String("user_id", req.UserId), zap.Float64("balance", balance))
	return &proto.GetBalanceResponse{Balance: balance}, nil
}

func (h *PaymentHandler) Refund(ctx context.Context, req *proto.RefundRequest) (*proto.RefundResponse, error) {
	// Check if transaction exists (simplified check)
	userId := "user123" // TODO: Get from transaction lookup

	err := h.repo.Refund(req.TransactionId, req.Amount)
	if err != nil {
		h.logger.Error("Refund error", zap.Error(err))
		return &proto.RefundResponse{Success: false, Error: "Refund failed"}, nil
	}

	// Update Redis balance
	newBalance, err := h.redis.IncrByFloat(ctx, "wallet:"+userId, req.Amount).Result()
	if err != nil {
		h.logger.Error("Redis update error", zap.Error(err))
	}

	h.logger.Info("Refund processed", zap.String("transaction_id", req.TransactionId), zap.Float64("amount", req.Amount))
	return &proto.RefundResponse{Success: true, NewBalance: newBalance}, nil
}
