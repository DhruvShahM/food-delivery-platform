package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"food-delivery-platform/services/order-service/internal/handler"
	"food-delivery-platform/services/order-service/internal/proto"
	"food-delivery-platform/services/order-service/internal/repository"
	orderproto "food-delivery-platform/common/proto"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"
)

type OrderTestSuite struct {
	suite.Suite
	db      *sql.DB
	repo    *repository.OrderRepo
	handler *handler.OrderHandler
	logger  *zap.Logger
	kafka   *kafkago.Writer
}

func (suite *OrderTestSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	var err error
	suite.db, err = sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	suite.NoError(err, "Failed to connect to database")
	
	err = suite.db.Ping()
	if err != nil {
		suite.T().Skip("Database not available, skipping tests")
		return
	}
	
	suite.kafka = kafkago.NewWriter(kafkago.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders",
	})
	suite.repo = repository.NewOrderRepo(suite.db, suite.kafka, suite.logger)
	suite.repo.Init()
	suite.handler = handler.NewOrderHandler(suite.repo, suite.logger)
}

func (suite *OrderTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
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
		UserId:       "user1",
		RestaurantId: "rest1",
		Items: []*orderproto.MenuItem{
			{Name: "Pizza", Price: 15.99},
			{Name: "Coke", Price: 2.99},
		},
	}
	resp, err := suite.handler.PlaceOrder(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.OrderId)
}

func (suite *OrderTestSuite) TestOrderHandler_GetOrders() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	// Place an order first
	placeReq := &proto.PlaceOrderRequest{
		UserId:       "user1",
		RestaurantId: "rest1",
		Items: []*orderproto.MenuItem{{Name: "Pizza", Price: 15.99}},
	}
	placeResp, _ := suite.handler.PlaceOrder(context.Background(), placeReq)
	
	req := &proto.GetOrdersRequest{UserId: "user1"}
	resp, err := suite.handler.GetOrders(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.Orders)
	
	// Verify the order is in the response
	found := false
	for _, order := range resp.Orders {
		if order.Id == placeResp.OrderId {
			found = true
			suite.Equal("created", order.Status) // Changed from "placed" to "created"
			break
		}
	}
	suite.True(found)
}

func (suite *OrderTestSuite) TestOrderHandler_UpdateStatus() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	// Place an order first
	placeReq := &proto.PlaceOrderRequest{
		UserId:       "user1",
		RestaurantId: "rest1",
		Items: []*orderproto.MenuItem{{Name: "Pizza", Price: 15.99}},
	}
	placeResp, _ := suite.handler.PlaceOrder(context.Background(), placeReq)
	
	req := &proto.UpdateStatusRequest{
		OrderId: placeResp.OrderId,
		Status:  "confirmed",
	}
	resp, err := suite.handler.UpdateStatus(context.Background(), req)

	suite.NoError(err)
	suite.True(resp.Success)
	
	// Verify status was updated
	getReq := &proto.GetOrdersRequest{UserId: "user1"}
	getResp, _ := suite.handler.GetOrders(context.Background(), getReq)
	
	found := false
	for _, order := range getResp.Orders {
		if order.Id == placeResp.OrderId {
			found = true
			suite.Equal("confirmed", order.Status)
			break
		}
	}
	suite.True(found)
}

func BenchmarkOrderRepo_GetOrders(b *testing.B) {
	suite := new(OrderTestSuite)
	suite.SetT(&testing.T{})
	suite.SetupTest()
	defer suite.TearDownTest()
	
	// Create test data
	userID := "bench_user"
	for i := 0; i < 10; i++ {
		orderID := fmt.Sprintf("bench_order_%d", i)
		restaurantID := "bench_rest"
		amount := 25.0
		err := suite.repo.CreateOrder(orderID, userID, restaurantID, amount)
		if err != nil {
			b.Fatalf("Failed to create test order: %v", err)
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orders, err := suite.repo.GetOrdersByUserID(userID)
		if err != nil {
			b.Fatalf("GetOrders failed: %v", err)
		}
		_ = orders // Prevent optimization
	}
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}