package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"food-delivery-platform/services/order-service/internal/handler"
	"food-delivery-platform/services/order-service/internal/proto"
	"food-delivery-platform/services/order-service/internal/repository"
	commonproto "food-delivery-platform/common/proto"
	"github.com/segmentio/kafka-go"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type OrderTestSuite struct {
	suite.Suite
	db      *sql.DB
	repo    *repository.OrderRepo
	handler *handler.OrderHandler
	kafka   *kafka.Writer
	logger  *zap.Logger
}

func (suite *OrderTestSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	
	// Kafka writer for testing
	suite.kafka = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "order-events",
		Balancer: &kafka.LeastBytes{},
	}
	
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

	suite.repo = repository.NewOrderRepo(suite.db, suite.kafka, suite.logger)
	suite.handler = handler.NewOrderHandler(suite.repo, suite.logger)
}

func (suite *OrderTestSuite) TearDownTest() {
	if suite.kafka != nil {
		suite.kafka.Close()
	}
}

func (suite *OrderTestSuite) TestOrderHandler_PlaceOrder() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.PlaceOrderRequest{
		UserId:      "test-user",
		RestaurantId: "test-restaurant",
		Items: []*commonproto.MenuItem{
			{Id: "1", Name: "Test Item", Price: 10.0, Available: true},
		},
	}

	resp, err := suite.handler.PlaceOrder(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotEmpty(resp.OrderId)
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}

// Helper function for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}