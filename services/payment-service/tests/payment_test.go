package tests

import (
	"context"
	"database/sql"
	"testing"

	"food-delivery-platform/services/payment-service/internal/handler"
	"food-delivery-platform/services/payment-service/internal/proto"
	"food-delivery-platform/services/payment-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"
)

type PaymentTestSuite struct {
	suite.Suite
	db      *sql.DB
	redis   *redis.Client
	repo    *repository.PaymentRepo
	handler *handler.PaymentHandler
	logger  *zap.Logger
}

func (suite *PaymentTestSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	var err error
	suite.db, err = sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	suite.NoError(err, "Failed to connect to database")
	
	err = suite.db.Ping()
	if err != nil {
		suite.T().Skip("Database not available, skipping tests")
		return
	}
	
	suite.redis = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	suite.repo = repository.NewPaymentRepo(suite.db, suite.logger)
	suite.repo.Init()
	suite.handler = handler.NewPaymentHandler(suite.repo, suite.redis, suite.logger)
}

func (suite *PaymentTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redis != nil {
		suite.redis.Close()
	}
}

func (suite *PaymentTestSuite) TestPaymentHandler_ProcessPayment_Success() {
	if suite.db == nil || suite.redis == nil {
		suite.T().Skip("Database or Redis not available")
		return
	}
	
	// Use the correct hardcoded user ID from the handler
	userId := "user123"
	
	// Setup wallet balance as integer (Redis expects numeric values for arithmetic operations)
	suite.redis.Set(context.Background(), "wallet:"+userId, 100, 0)
	
	req := &proto.ProcessPaymentRequest{
		OrderId: "order123",
		Amount:  20.0,
	}
	resp, err := suite.handler.ProcessPayment(context.Background(), req)

	suite.NoError(err)
	suite.Equal("completed", resp.Status)
	suite.NotEmpty(resp.PaymentId)
	suite.Empty(resp.Error)
	
	// Verify balance updated
	balance, _ := suite.redis.Get(context.Background(), "wallet:"+userId).Int64()
	suite.Equal(int64(80), balance)
}

func (suite *PaymentTestSuite) TestPaymentHandler_ProcessPayment_InsufficientFunds() {
	if suite.db == nil || suite.redis == nil {
		suite.T().Skip("Database or Redis not available")
		return
	}
	
	userId := "user123"
	
	// Setup low balance
	suite.redis.Set(context.Background(), "wallet:"+userId, 10, 0)
	
	req := &proto.ProcessPaymentRequest{
		OrderId: "order456",
		Amount:  50.0,
	}
	resp, err := suite.handler.ProcessPayment(context.Background(), req)

	suite.NoError(err)
	suite.Equal("failed", resp.Status)
	suite.NotEmpty(resp.Error)
	suite.Empty(resp.PaymentId)
	
	// Verify balance unchanged
	balance, _ := suite.redis.Get(context.Background(), "wallet:"+userId).Int64()
	suite.Equal(int64(10), balance)
}

func (suite *PaymentTestSuite) TestPaymentHandler_GetBalance() {
	if suite.redis == nil {
		suite.T().Skip("Redis not available")
		return
	}
	
	userId := "user123"
	suite.redis.Set(context.Background(), "wallet:"+userId, 75.50, 0)
	
	req := &proto.GetBalanceRequest{UserId: userId}
	resp, err := suite.handler.GetBalance(context.Background(), req)

	suite.NoError(err)
	suite.Equal(75.50, resp.Balance)
	suite.Empty(resp.Error)
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}