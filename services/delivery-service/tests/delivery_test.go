package tests

import (
	"context"
	"database/sql"
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
	var err error
	suite.db, err = sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	suite.NoError(err, "Failed to connect to database")

	err = suite.db.Ping()
	if err != nil {
		suite.T().Skip("Database not available, skipping tests")
		return
	}

	suite.repo = repository.NewDeliveryRepo(suite.db, suite.logger)
	suite.repo.Init()
	suite.handler = handler.NewDeliveryHandler(suite.repo, suite.logger)
}

func (suite *DeliveryTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *DeliveryTestSuite) TestDeliveryHandler_AssignDelivery() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	req := &proto.AssignDeliveryRequest{OrderId: "order1"}
	resp, err := suite.handler.AssignDelivery(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.AgentId)
}

func (suite *DeliveryTestSuite) TestDeliveryHandler_UpdateStatus() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// First assign a delivery
	assignReq := &proto.AssignDeliveryRequest{OrderId: "order123"}
	assignResp, _ := suite.handler.AssignDelivery(context.Background(), assignReq)

	req := &proto.UpdateStatusRequest{
		DeliveryId: assignResp.AgentId,
		Status:     "picked_up",
	}
	resp, err := suite.handler.UpdateStatus(context.Background(), req)

	suite.NoError(err)
	suite.True(resp.Success)
}

func TestDeliveryTestSuite(t *testing.T) {
	suite.Run(t, new(DeliveryTestSuite))
}
