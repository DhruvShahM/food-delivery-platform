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
	userId := req.UserId // Ensure proto has userid; if using OrderId, extract user from context/JWT
	key := "wallet:" + userId
	balance, err := h.redis.Get(ctx, key).Float64()
	if err != nil {
		// Fall back to DB if Redis miss
		balance, err = h.repo.GetBalance(userId)
		if err != nil {
			h.logger.Error("Get balance error", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal error")
		}
		// Cache in Redis for future
		h.redis.Set(ctx, key, balance, 0)
	}

	if balance < req.Amount {
		h.logger.Warn("Insufficient balance",
			zap.String("userid", userId),
			zap.Float64("amount", req.Amount),
			zap.Float64("balance", balance))
		return &proto.ProcessPaymentResponse{
			Status: "failed",
			Error:  "Insufficient balance",
		}, nil
	}

	// Process payment in DB
	err = h.repo.Process(userId, req.Amount)
	if err != nil {
		h.logger.Error("Process payment error", zap.Error(err))
		return &proto.ProcessPaymentResponse{
			Status: "failed",
			Error:  "Internal error",
		}, nil
	}

	// Update Redis balance
	_, err = h.redis.DecrBy(ctx, key, req.Amount).Result()
	if err != nil {
		h.logger.Error("Redis update error", zap.Error(err))
	}

	paymentId := fmt.Sprintf("payment_%d", time.Now().UnixNano())
	h.logger.Info("Payment processed",
		zap.String("paymentid", paymentId),
		zap.String("userid", userId),
		zap.Float64("amount", req.Amount))
	return &proto.ProcessPaymentResponse{
		PaymentId: paymentId,
		Status:    "completed",
	}, nil
}

func (h *PaymentHandler) GetBalance(ctx context.Context, req *proto.GetBalanceRequest) (*proto.GetBalanceResponse, error) {
	balance, err := h.redis.Get(ctx, "wallet:"+req.UserId).Float64()
	if err != nil {
		balance, err = h.repo.GetBalance(req.UserId)
		if err != nil {
			h.logger.Error("Get balance error", zap.Error(err))
			return &proto.GetBalanceResponse{Error: "Internal error"}, nil
		}
		h.redis.Set(ctx, "wallet:"+req.UserId, balance, 0)
	}
	h.logger.Info("Balance retrieved", zap.String("userid", req.UserId), zap.Float64("balance", balance))
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
