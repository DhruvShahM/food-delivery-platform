package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"food-delivery-platform/services/delivery-service/internal/handler"
	"food-delivery-platform/services/delivery-service/internal/proto"
	"food-delivery-platform/services/delivery-service/internal/repository"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type DeliveryTestSuite struct {
	suite.Suite
	db      *sql.DB
	repo    *repository.DeliveryRepo
	handler *handler.DeliveryHandler
	logger  *zap.Logger
}

func (suite *DeliveryTestSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	
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

	suite.repo = repository.NewDeliveryRepo(suite.db, suite.logger)
	suite.handler = handler.NewDeliveryHandler(suite.repo, suite.logger)
}

func (suite *DeliveryTestSuite) TestDeliveryHandler_AssignDelivery() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.AssignDeliveryRequest{
		OrderId: "test-order",
	}

	resp, err := suite.handler.AssignDelivery(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotEmpty(resp.AgentId)
}

func (suite *DeliveryTestSuite) TestDeliveryHandler_UpdateStatus() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.UpdateStatusRequest{
		DeliveryId: "test-delivery",
		Status:     "delivered",
	}

	resp, err := suite.handler.UpdateStatus(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)
}

func TestDeliveryTestSuite(t *testing.T) {
	suite.Run(t, new(DeliveryTestSuite))
}

// Helper function for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
