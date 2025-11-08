package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"food-delivery-platform/services/auth-service/internal/handler"
	"food-delivery-platform/services/auth-service/internal/proto"
	"food-delivery-platform/services/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthTestSuite struct {
	suite.Suite
	db      *sql.DB
	repo    *repository.UserRepo
	handler *handler.AuthHandler
	logger  *zap.Logger
}

func (suite *AuthTestSuite) SetupTest() {
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

	suite.repo = repository.NewUserRepo(suite.db, suite.logger)
}

func (suite *AuthTestSuite) TestAuthHandler_Login_Success() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create unique test user for this test
	testEmail := fmt.Sprintf("test_success_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "password")
	suite.NoError(err)

	// Test login
	req := &proto.LoginRequest{
		Email:    testEmail,
		Password: "password",
	}

	suite.handler = handler.NewAuthHandler(suite.repo, "test-secret", suite.logger)
	resp, err := suite.handler.Login(context.Background(), req)
	suite.NoError(err)
	suite.NotEmpty(resp.Token)

	// Verify JWT token
	token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	suite.NoError(err)
	suite.True(token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	suite.True(ok)
	suite.Equal(testEmail, claims["email"])
}

func (suite *AuthTestSuite) TestAuthHandler_Login_InvalidPassword() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create unique test user for this test
	testEmail := fmt.Sprintf("test_invalid_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "correct_password")
	suite.NoError(err)

	// Test login with wrong password
	req := &proto.LoginRequest{
		Email:    testEmail,
		Password: "wrong_password",
	}

	suite.handler = handler.NewAuthHandler(suite.repo, "test-secret", suite.logger)
	_, err = suite.handler.Login(context.Background(), req)
	suite.Error(err)
}

func (suite *AuthTestSuite) TestAuthHandler_Login_UserNotFound() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Test login with non-existent user
	req := &proto.LoginRequest{
		Email:    "nonexistent@test.com",
		Password: "password",
	}

	suite.handler = handler.NewAuthHandler(suite.repo, "test-secret", suite.logger)
	_, err := suite.handler.Login(context.Background(), req)
	suite.Error(err)
}

func (suite *AuthTestSuite) TestAuthHandler_Register_Success() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Test registration
	testEmail := fmt.Sprintf("register_success_%d@test.com", time.Now().UnixNano())
	req := &proto.RegisterRequest{
		Email:    testEmail,
		Password: "password",
	}

	suite.handler = handler.NewAuthHandler(suite.repo, "test-secret", suite.logger)
	resp, err := suite.handler.Register(context.Background(), req)
	suite.NoError(err)
	suite.NotEmpty(resp.Token)

	// Verify user was created in database
	user, err := suite.repo.GetUserByEmail(testEmail)
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(testEmail, user.Email)
}

func (suite *AuthTestSuite) TestAuthHandler_Register_DuplicateEmail() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create user first
	testEmail := fmt.Sprintf("duplicate_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "password")
	suite.NoError(err)

	// Try to register with same email
	req := &proto.RegisterRequest{
		Email:    testEmail,
		Password: "password2",
	}

	suite.handler = handler.NewAuthHandler(suite.repo, "test-secret", suite.logger)
	_, err = suite.handler.Register(context.Background(), req)
	suite.Error(err)
}

func (suite *AuthTestSuite) TestUserRepo_GetUserByEmail() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Create test user
	testEmail := fmt.Sprintf("repo_test_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "password")
	suite.NoError(err)

	// Test GetUserByEmail
	user, err := suite.repo.GetUserByEmail(testEmail)
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(testEmail, user.Email)

	// Test non-existent user
	user, err = suite.repo.GetUserByEmail("nonexistent@test.com")
	suite.Error(err)
	suite.Nil(user)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

// Helper function for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}