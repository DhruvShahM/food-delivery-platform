package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"food-delivery-platform/services/auth-service/internal/handler"
	"food-delivery-platform/services/auth-service/internal/proto"
	"food-delivery-platform/services/auth-service/internal/repository"
	"go.uber.org/zap"
	"github.com/stretchr/testify/suite"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
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
	var err error
	suite.db, err = sql.Open("postgres", "postgres://root:root@localhost:5432/fooddb?sslmode=disable")
	suite.NoError(err, "Failed to connect to database")
	
	err = suite.db.Ping()
	if err != nil {
		suite.T().Skip("Database not available, skipping tests")
		return
	}
	
	suite.repo = repository.NewUserRepo(suite.db, suite.logger)
	// Don't call Init() here to avoid conflicts between tests
	suite.handler = handler.NewAuthHandler(suite.repo, "testsecret", suite.logger)
}

func (suite *AuthTestSuite) TearDownTest() {
	if suite.db != nil {
		// Clean up test data
		suite.db.Exec("DELETE FROM users WHERE email LIKE 'test%@%' OR email LIKE 'newuser%@%' OR email LIKE 'duplicate%@%' OR email LIKE 'repo%@%'")
		suite.db.Close()
	}
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

	req := &proto.LoginRequest{Email: testEmail, Password: "password"}
	resp, err := suite.handler.Login(context.Background(), req)

	suite.NoError(err)
	suite.Empty(resp.Error)
	suite.NotEmpty(resp.Token)
	
	token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("testsecret"), nil
	})
	suite.NoError(err)
	suite.True(token.Valid)
}

func (suite *AuthTestSuite) TestAuthHandler_Login_InvalidPassword() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	testEmail := fmt.Sprintf("test_invalid_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "password")
	suite.NoError(err)

	req := &proto.LoginRequest{Email: testEmail, Password: "wrongpassword"}
	resp, err := suite.handler.Login(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.Error)
	suite.Empty(resp.Token)
}

func (suite *AuthTestSuite) TestAuthHandler_Login_UserNotFound() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	req := &proto.LoginRequest{Email: "nonexistent@test.com", Password: "password"}
	resp, err := suite.handler.Login(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.Error)
	suite.Empty(resp.Token)
}

func (suite *AuthTestSuite) TestAuthHandler_Register_Success() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	testEmail := fmt.Sprintf("register_success_%d@test.com", time.Now().UnixNano())
	req := &proto.RegisterRequest{
		Email:    testEmail,
		Password: "password123",
	}
	resp, err := suite.handler.Register(context.Background(), req)

	suite.NoError(err)
	suite.Empty(resp.Error)
	suite.NotEmpty(resp.Token)
}

func (suite *AuthTestSuite) TestAuthHandler_Register_DuplicateEmail() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	testEmail := fmt.Sprintf("duplicate_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "password")
	suite.NoError(err)

	req := &proto.RegisterRequest{
		Email:    testEmail,
		Password: "password123",
	}
	resp, err := suite.handler.Register(context.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(resp.Error)
	suite.Empty(resp.Token)
}

func (suite *AuthTestSuite) TestUserRepo_GetUserByEmail() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	
	testEmail := fmt.Sprintf("repo_test_%d@test.com", time.Now().UnixNano())
	err := suite.repo.CreateUser(testEmail, "pass")
	suite.NoError(err)

	user, err := suite.repo.GetUserByEmail(testEmail)
	suite.NoError(err)
	suite.Equal(testEmail, user.Email)
	suite.NotEmpty(user.Password)

	_, err = suite.repo.GetUserByEmail("missing@test.com")
	suite.Error(err)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}