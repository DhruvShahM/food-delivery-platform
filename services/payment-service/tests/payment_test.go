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

func TestPaymentHandler_ProcessPayment_Success(t *testing.T) {
	logger := zap.NewDevelopment()
	db, err := sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()
	repo := repository.NewPaymentRepo(db, logger)
	require.NoError(t, repo.Init())

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer redisClient.Close()

	// Set initial balance
	ctx := context.Background()
	require.NoError(t, redisClient.Set(ctx, "wallet:user1", 100.0, 0).Err())

	h := handler.NewPaymentHandler(repo, redisClient, logger)
	req := &proto.ProcessPaymentRequest{UserId: "user1", Amount: 20.0} // Use UserId
	resp, err := h.ProcessPayment(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.PaymentId)
	require.Equal(t, "completed", resp.Status) // Assert on Status

	// Verify balance deducted
	balance, _ := redisClient.Get(ctx, "wallet:user1").Float64()
	require.Equal(t, 80.0, balance)
}

func TestPaymentHandler_ProcessPayment_InsufficientFunds(t *testing.T) {
	// Similar setup...
	req := &proto.ProcessPaymentRequest{UserId: "user1", Amount: 200.0}
	resp, err := h.ProcessPayment(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "failed", resp.Status)
	require.Equal(t, "Insufficient balance", resp.Error)
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
