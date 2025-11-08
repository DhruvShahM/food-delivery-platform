// services/order-service/internal/client/payment_client.go
package client

import (
	"context"
	"time"

	"food-delivery-platform/services/order-service/internal/middleware"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PaymentClient struct {
	client payment.PaymentServiceClient
	cb     *middleware.CircuitBreaker
	logger *zap.Logger
}

func NewPaymentClient(conn *grpc.ClientConn, cbManager *middleware.CircuitBreakerManager, logger *zap.Logger) *PaymentClient {
	cb := cbManager.GetOrCreateBreaker("payment-client", &middleware.CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     30 * time.Second,
		Timeout:      10 * time.Second,
		FailureRatio: 0.5,
		Logger:       logger,
	})

	return &PaymentClient{
		client: payment.NewPaymentServiceClient(conn),
		cb:     cb,
		logger: logger,
	}
}

func (pc *PaymentClient) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	result, err := pc.cb.cb.Execute(func() (interface{}, error) {
		return pc.client.ProcessPayment(ctx, req)
	})

	if err != nil {
		pc.logger.Error("Payment service call failed",
			zap.Error(err),
			zap.String("method", "ProcessPayment"),
		)
		return nil, err
	}

	return result.(*payment.ProcessPaymentResponse), nil
}