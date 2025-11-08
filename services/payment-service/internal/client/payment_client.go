// services/payment-service/internal/client/payment_client.go
package client

import (
	"context"
	"time"

	payment "food-delivery-platform/services/payment-service/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PaymentClient struct {
	client payment.PaymentServiceClient
	logger *zap.Logger
}

func NewPaymentClient(conn *grpc.ClientConn, logger *zap.Logger) *PaymentClient {
	return &PaymentClient{
		client: payment.NewPaymentServiceClient(conn),
		logger: logger,
	}
}

func (pc *PaymentClient) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := pc.client.ProcessPayment(ctx, req)
	if err != nil {
		pc.logger.Error("Payment service call failed",
			zap.Error(err),
			zap.String("method", "ProcessPayment"),
		)
		return nil, err
	}

	return resp, nil
}

func (pc *PaymentClient) GetBalance(ctx context.Context, req *payment.GetBalanceRequest) (*payment.GetBalanceResponse, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := pc.client.GetBalance(ctx, req)
	if err != nil {
		pc.logger.Error("Payment service call failed",
			zap.Error(err),
			zap.String("method", "GetBalance"),
		)
		return nil, err
	}

	return resp, nil
}
