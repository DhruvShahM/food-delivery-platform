package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
}

// HTTP rate limiting config
type RateLimitConfig struct {
	Requests int                       // Number of requests allowed
	Window   time.Duration             // Time window
	KeyFunc  func(*gin.Context) string // Function to generate rate limit key
}

// gRPC rate limiting config
type GRPCRateLimitConfig struct {
	Requests int                                                 // Number of requests allowed
	Window   time.Duration                                       // Time window
	KeyFunc  func(context.Context, *grpc.UnaryServerInfo) string // Function to generate rate limit key
}

func NewRateLimiter(redis *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		logger: logger,
	}
}

// Allow checks if the request should be allowed based on rate limiting rules
func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) (bool, int, error) {
	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Clean up old entries and add new entry in a pipeline
	pipe := rl.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
	pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := results[2].(*redis.IntCmd).Val()
	return count <= int64(limit), int(count), nil
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		allowed, currentCount, err := rl.Allow(key, config.Requests, config.Window)
		if err != nil {
			rl.logger.Error("Rate limiter error", zap.Error(err))
			c.Next() // Allow request on error to avoid blocking legitimate traffic
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, config.Requests-currentCount)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Window).Unix()))

		if !allowed {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.Int("current", currentCount),
				zap.Int("limit", config.Requests))

			c.Header("Retry-After", fmt.Sprintf("%d", int(config.Window.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(config.Window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Predefined rate limit configurations
func (rl *RateLimiter) IPRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return rl.RateLimitMiddleware(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())
		},
	})
}

func (rl *RateLimiter) UserRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return rl.RateLimitMiddleware(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc: func(c *gin.Context) string {
			userID, exists := c.Get("user_id")
			if !exists {
				return fmt.Sprintf("ratelimit:anon:%s", c.ClientIP())
			}
			return fmt.Sprintf("ratelimit:user:%s", userID.(string))
		},
	})
}

func (rl *RateLimiter) EndpointRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return rl.RateLimitMiddleware(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("ratelimit:endpoint:%s:%s", c.Request.Method, c.Request.URL.Path)
		},
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GRPCRateLimitInterceptor creates a gRPC interceptor for rate limiting
func (rl *RateLimiter) GRPCRateLimitInterceptor(config GRPCRateLimitConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		key := config.KeyFunc(ctx, info)

		allowed, currentCount, err := rl.Allow(key, config.Requests, config.Window)
		if err != nil {
			rl.logger.Error("gRPC rate limiter error", zap.Error(err))
			// Allow request on error to avoid blocking legitimate traffic
			return handler(ctx, req)
		}

		if !allowed {
			rl.logger.Warn("gRPC rate limit exceeded",
				zap.String("key", key),
				zap.Int("current", currentCount),
				zap.Int("limit", config.Requests),
				zap.String("method", info.FullMethod))

			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// Predefined gRPC rate limit configurations
func (rl *RateLimiter) GRPCIPRateLimit(requests int, window time.Duration) grpc.UnaryServerInterceptor {
	return rl.GRPCRateLimitInterceptor(GRPCRateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc: func(ctx context.Context, info *grpc.UnaryServerInfo) string {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return "ratelimit:grpc:unknown"
			}

			peer := md[":authority"]
			if len(peer) > 0 {
				return fmt.Sprintf("ratelimit:grpc:ip:%s", peer[0])
			}

			// Fallback to method-based limiting
			return fmt.Sprintf("ratelimit:grpc:method:%s", info.FullMethod)
		},
	})
}

func (rl *RateLimiter) GRPCUserRateLimit(requests int, window time.Duration) grpc.UnaryServerInterceptor {
	return rl.GRPCRateLimitInterceptor(GRPCRateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc: func(ctx context.Context, info *grpc.UnaryServerInfo) string {
			userID, ok := ctx.Value("user_id").(string)
			if !ok {
				return fmt.Sprintf("ratelimit:grpc:anon:%s", info.FullMethod)
			}
			return fmt.Sprintf("ratelimit:grpc:user:%s", userID)
		},
	})
}
