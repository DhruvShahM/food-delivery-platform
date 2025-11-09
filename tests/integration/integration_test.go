package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	authPublic "food-delivery-platform/services/auth-service/public"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnvOrDefault("DB_USER", "root"),
			getEnvOrDefault("DB_PASSWORD", "root"),
			getEnvOrDefault("DB_HOST", "localhost"),
			getEnvOrDefault("DB_PORT", "5432"),
			getEnvOrDefault("DB_NAME", "fooddb"))
	}
	suite.db, _ = sql.Open("postgres", dbURL)
	// Ensure DB is reachable (retry to handle startup races)
	for i := 0; i < 20; i++ {
		if err := suite.db.Ping(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnvOrDefault("DB_USER", "postgres"),
			getEnvOrDefault("DB_PASSWORD", "postgres"),
			getEnvOrDefault("DB_HOST", "localhost"),
			getEnvOrDefault("DB_PORT", "5432"),
			getEnvOrDefault("DB_NAME", "fooddb"))
	}
	db, _ := sql.Open("postgres", dbURL)
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
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("RUN_INTEGRATION_TESTS is not set; skipping integration tests")
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}

// Helper to read env var with default
func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
