package tests

import (
	"context"
	"database/sql"
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
	redis   *redis.Client
	repo    *repository.MenuRepo
	handler *handler.RestaurantHandler
	logger  *zap.Logger
}

func (suite *RestaurantTestSuite) SetupTest() {
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
	suite.repo = repository.NewMenuRepo(suite.db, suite.logger)
	suite.repo.Init()
	suite.handler = handler.NewRestaurantHandler(suite.repo, suite.redis, suite.logger)
}

func (suite *RestaurantTestSuite) TearDownTest() {
	if suite.db != nil {
		// Clean up test data
		suite.db.Exec("DELETE FROM menu WHERE restaurant_id = 'rest1' AND id LIKE 'item%'")
		suite.db.Close()
	}
	if suite.redis != nil {
		suite.redis.Close()
	}
}

func (suite *RestaurantTestSuite) TestRestaurantHandler_GetMenu() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create test data using the correct method signature
	err := suite.repo.CreateMenuItem("rest1", "test_item_1", "Pizza", 15.99)
	suite.NoError(err)

	req := &proto.GetMenuRequest{RestaurantId: "rest1"}
	resp, err := suite.handler.GetMenu(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.Items)
	// Don't check exact count since Init() adds sample data
}

func (suite *RestaurantTestSuite) TestMenuRepo_UpdateAvailability() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create item first
	err := suite.repo.CreateMenuItem("rest1", "test_item_2", "Salad", 8.99)
	suite.NoError(err)

	err = suite.repo.UpdateAvailability("test_item_2", false)
	suite.NoError(err)
}

func (suite *RestaurantTestSuite) TestMenuRepo_GetMenuItems() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create test items
	err := suite.repo.CreateMenuItem("rest1", "test_item_3", "Pizza", 15.99)
	suite.NoError(err)
	err = suite.repo.CreateMenuItem("rest1", "test_item_4", "Burger", 12.99)
	suite.NoError(err)

	items, err := suite.repo.GetMenu("rest1")
	suite.NoError(err)
	suite.True(len(items) >= 2) // At least our test items plus any sample data
}

func TestRestaurantTestSuite(t *testing.T) {
	suite.Run(t, new(RestaurantTestSuite))
}
