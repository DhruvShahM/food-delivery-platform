package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	authPublic "food-delivery-platform/services/auth-service/public"
	
	"go.uber.org/zap"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"
)

// IntegrationTestSuite tests end-to-end service interactions
type IntegrationTestSuite struct {
	suite.Suite
	db          *sql.DB
	authService authPublic.AuthService
	logger      *zap.Logger
	testEmail   string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.logger, _ = zap.NewDevelopment()
	suite.db, _ = sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	suite.authService = authPublic.NewAuthService(suite.db, suite.logger)
	// Use unique email for each test run
	suite.testEmail = fmt.Sprintf("integration_test_%d@example.com", time.Now().UnixNano())
}

func (suite *IntegrationTestSuite) TestAuthService() {
	// Test registration
	req := authPublic.NewRegisterRequest()
	req.Email = suite.testEmail
	req.Password = "password123"
	
	resp, err := suite.authService.Register(context.Background(), req)
	suite.NoError(err)
	suite.NotEmpty(resp.Token)
	
	// Test login
	loginReq := authPublic.NewLoginRequest()
	loginReq.Email = suite.testEmail
	loginReq.Password = "password123"
	
	loginResp, err := suite.authService.Login(context.Background(), loginReq)
	suite.NoError(err)
	suite.NotEmpty(loginResp.Token)
}

func BenchmarkAuthService(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	db, _ := sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	authService := authPublic.NewAuthService(db, logger)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("bench_user_%d_%d@example.com", i, time.Now().UnixNano())
		
		req := authPublic.NewRegisterRequest()
		req.Email = email
		req.Password = "password123"
		
		resp, err := authService.Register(context.Background(), req)
		if err != nil || resp.Token == "" {
			b.Fatalf("Registration failed: %v", err)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
