package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"food-delivery-platform/services/payment-service/internal/client"
	"food-delivery-platform/services/payment-service/internal/handler"
	"food-delivery-platform/services/payment-service/internal/proto"
	"food-delivery-platform/services/payment-service/internal/repository"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type PaymentTestSuite struct {
	suite.Suite
	db      *sql.DB
	repo    *repository.PaymentRepo
	handler *handler.PaymentHandler
	redis   *redis.Client
	client  *client.PaymentClient
	logger  *zap.Logger
}

func (suite *PaymentTestSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	suite.redis = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	suite.client = client.NewPaymentClient(nil, suite.logger) // Remove cbManager parameter

	// Set environment variables for testing
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "fooddb")

	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnvOrDefault("DB_USER", "postgres"),
			getEnvOrDefault("DB_PASSWORD", "postgres"),
			getEnvOrDefault("DB_HOST", "localhost"),
			getEnvOrDefault("DB_PORT", "5432"),
			getEnvOrDefault("DB_NAME", "fooddb"))
	}

	suite.db, err = sql.Open("postgres", dbURL)
	suite.NoError(err, "Failed to connect to database")

	err = suite.db.Ping()
	if err != nil {
		suite.T().Skip("Database not available, skipping tests")
		return
	}

	suite.repo = repository.NewPaymentRepo(suite.db, suite.logger)
	suite.handler = handler.NewPaymentHandler(suite.repo, suite.redis, suite.logger)
}

func (suite *PaymentTestSuite) TearDownTest() {
	if suite.redis != nil {
		suite.redis.Close()
	}
}

func (suite *PaymentTestSuite) TestPaymentHandler_GetBalance() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.GetBalanceRequest{UserId: "test-user"}
	resp, err := suite.handler.GetBalance(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *PaymentTestSuite) TestPaymentHandler_ProcessPayment_Success() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.ProcessPaymentRequest{
		OrderId: "test-order", // Changed from UserId to OrderId
		Amount:  100.0,
	}

	resp, err := suite.handler.ProcessPayment(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotEmpty(resp.PaymentId)        // Changed from Success to PaymentId
	suite.Equal("completed", resp.Status) // Check status field
}

func (suite *PaymentTestSuite) TestPaymentHandler_ProcessPayment_InsufficientFunds() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.ProcessPaymentRequest{
		OrderId: "test-order", // Changed from UserId to OrderId
		Amount:  1000000.0,    // Very large amount
	}

	resp, err := suite.handler.ProcessPayment(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal("failed", resp.Status) // Changed from !resp.Success to status check
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}

// Helper function for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
