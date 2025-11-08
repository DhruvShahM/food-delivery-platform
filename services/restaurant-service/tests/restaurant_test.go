package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"food-delivery-platform/services/restaurant-service/internal/handler"
	"food-delivery-platform/services/restaurant-service/internal/proto"
	"food-delivery-platform/services/restaurant-service/internal/repository"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type RestaurantTestSuite struct {
	suite.Suite
	db      *sql.DB
	repo    *repository.MenuRepo
	handler *handler.RestaurantHandler
	redis   *redis.Client
	logger  *zap.Logger
}

func (suite *RestaurantTestSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	suite.redis = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	
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

	suite.repo = repository.NewMenuRepo(suite.db, suite.logger)
	suite.handler = handler.NewRestaurantHandler(suite.repo, suite.redis, suite.logger)
}

func (suite *RestaurantTestSuite) TearDownTest() {
	if suite.redis != nil {
		suite.redis.Close()
	}
}

func (suite *RestaurantTestSuite) TestMenuRepo_UpdateAvailability() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Test updating item availability
	err := suite.repo.UpdateAvailability("1", false)
	suite.NoError(err)
}

func (suite *RestaurantTestSuite) TestRestaurantHandler_GetMenu() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.GetMenuRequest{}
	resp, err := suite.handler.GetMenu(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
}

func TestRestaurantTestSuite(t *testing.T) {
	suite.Run(t, new(RestaurantTestSuite))
}

// Helper function for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
