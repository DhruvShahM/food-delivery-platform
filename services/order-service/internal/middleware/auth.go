package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthConfig struct {
	JWTSecret     string
	RedisClient   *redis.Client
	TokenExpiry   time.Duration
	Logger        *zap.Logger
}

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	config *AuthConfig
}

func NewAuthService(config *AuthConfig) *AuthService {
	return &AuthService{config: config}
}

// GenerateToken creates a new JWT token for a user
func (a *AuthService) GenerateToken(userID, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "food-delivery-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}

// ValidateToken validates a JWT token
func (a *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// BlacklistToken adds a token to the blacklist
func (a *AuthService) BlacklistToken(tokenString string, expiry time.Duration) error {
	ctx := context.Background()
	return a.config.RedisClient.Set(ctx, "blacklist:"+tokenString, "true", expiry).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (a *AuthService) IsTokenBlacklisted(tokenString string) bool {
	ctx := context.Background()
	result, err := a.config.RedisClient.Get(ctx, "blacklist:"+tokenString).Result()
	return err == nil && result == "true"
}

// JWTMiddleware validates JWT tokens
func (a *AuthService) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Check if token is blacklisted
		if a.IsTokenBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			a.config.Logger.Error("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("token", tokenString)

		c.Next()
	}
}

// OptionalAuthMiddleware makes authentication optional (for public endpoints)
func (a *AuthService) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		if !a.IsTokenBlacklisted(tokenString) {
			if claims, err := a.ValidateToken(tokenString); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				c.Set("role", claims.Role)
				c.Set("token", tokenString)
			}
		}

		c.Next()
	}
}

// RoleBasedAuthMiddleware checks for specific roles
func (a *AuthService) RoleBasedAuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// LogoutHandler invalidates the current token
func (a *AuthService) LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, exists := c.Get("token")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No active token"})
			return
		}

		tokenString := token.(string)
		
		// Blacklist the token (assuming standard 24h expiry for simplicity)
		if err := a.BlacklistToken(tokenString, 24*time.Hour); err != nil {
			a.config.Logger.Error("Failed to blacklist token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

// GRPCAuthInterceptor validates JWT tokens for gRPC calls
func (a *AuthService) GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for health checks
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header required")
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.SplitN(authHeader[0], " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		tokenString := tokenParts[1]

		// Check if token is blacklisted
		if a.IsTokenBlacklisted(tokenString) {
			return nil, status.Error(codes.Unauthenticated, "token has been revoked")
		}

		// Validate token
		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			a.config.Logger.Error("gRPC token validation failed", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Add user information to context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)

		return handler(ctx, req)
	}
}

// GRPCRoleBasedInterceptor checks for specific roles in gRPC calls
func (a *AuthService) GRPCRoleBasedInterceptor(requiredRoles ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		role, ok := ctx.Value("role").(string)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "authentication required")
		}

		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				return handler(ctx, req)
			}
		}

		return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
	}
}
