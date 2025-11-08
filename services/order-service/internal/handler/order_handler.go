package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"food-delivery-platform/services/order-service/internal/middleware"
	"food-delivery-platform/services/order-service/internal/proto"
	"food-delivery-platform/services/order-service/internal/repository"
	commonproto "food-delivery-platform/common/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	proto.UnimplementedOrderServiceServer
	repo    *repository.OrderRepo
	logger  *zap.Logger
}

func NewOrderHandler(repo *repository.OrderRepo, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{repo: repo, logger: logger}
}

func (h *OrderHandler) PlaceOrder(ctx context.Context, req *proto.PlaceOrderRequest) (*proto.PlaceOrderResponse, error) {
	h.logger.Info("Place order", zap.String("user_id", req.UserId), zap.String("restaurant_id", req.RestaurantId))

	orderID := fmt.Sprintf("order_%d", time.Now().UnixNano())
	amount := 0.0
	for _, item := range req.Items {
		amount += item.Price
	}

	err := h.repo.CreateOrder(orderID, req.UserId, req.RestaurantId, amount)
	if err != nil {
		h.logger.Error("Create order error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	// Publish Kafka event
	evt := &commonproto.OrderEvent{
		OrderId:       orderID,
		UserId:        req.UserId,
		RestaurantId:  req.RestaurantId,
		Status:        "created",
		Amount:        amount,
		Timestamp:     time.Now().Format(time.RFC3339),
	}
	h.repo.PublishEvent(evt)

	h.logger.Info("Order created", zap.String("order_id", orderID))
	return &proto.PlaceOrderResponse{OrderId: orderID}, nil
}

func (h *OrderHandler) GetOrders(ctx context.Context, req *proto.GetOrdersRequest) (*proto.GetOrdersResponse, error) {
	h.logger.Info("Get orders", zap.String("user_id", req.UserId))

	orders, err := h.repo.GetOrdersByUserID(req.UserId)
	if err != nil {
		h.logger.Error("Get orders error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	h.logger.Info("Orders retrieved", zap.String("user_id", req.UserId), zap.Int("count", len(orders)))
	return &proto.GetOrdersResponse{Orders: orders}, nil
}

func (h *OrderHandler) UpdateStatus(ctx context.Context, req *proto.UpdateStatusRequest) (*proto.UpdateStatusResponse, error) {
	h.logger.Info("Update order status", zap.String("order_id", req.OrderId), zap.String("status", req.Status))

	err := h.repo.UpdateOrderStatus(req.OrderId, req.Status)
	if err != nil {
		h.logger.Error("Update order status error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	// Publish status update event
	evt := &commonproto.OrderEvent{
		OrderId:   req.OrderId,
		Status:    req.Status,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.repo.PublishEvent(evt)

	h.logger.Info("Order status updated", zap.String("order_id", req.OrderId), zap.String("status", req.Status))
	return &proto.UpdateStatusResponse{Success: true}, nil
}

// HTTP Handlers for REST API endpoints
func (h *OrderHandler) PlaceOrderHTTP(c *gin.Context) {
	var req struct {
		UserID       string                   `json:"user_id" binding:"required"`
		RestaurantID string                   `json:"restaurant_id" binding:"required"`
		Items        []commonproto.MenuItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID := fmt.Sprintf("order_%d", time.Now().UnixNano())
	amount := 0.0
	for _, item := range req.Items {
		amount += item.Price
	}

	err := h.repo.CreateOrder(orderID, req.UserID, req.RestaurantID, amount)
	if err != nil {
		h.logger.Error("Create order error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Publish Kafka event
	evt := &commonproto.OrderEvent{
		OrderId:       orderID,
		UserId:        req.UserID,
		RestaurantId:  req.RestaurantID,
		Status:        "created",
		Amount:        amount,
		Timestamp:     time.Now().Format(time.RFC3339),
	}
	h.repo.PublishEvent(evt)

	h.logger.Info("Order created via HTTP", zap.String("order_id", orderID))
	c.JSON(http.StatusCreated, gin.H{"order_id": orderID})
}

func (h *OrderHandler) GetOrdersHTTP(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orders, err := h.repo.GetOrdersByUserID(userID.(string))
	if err != nil {
		h.logger.Error("Get orders error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	h.logger.Info("Orders retrieved via HTTP", zap.String("user_id", userID.(string)), zap.Int("count", len(orders)))
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) UpdateOrderStatusHTTP(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.repo.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		h.logger.Error("Update order status error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Publish status update event
	evt := &commonproto.OrderEvent{
		OrderId:   orderID,
		Status:    req.Status,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.repo.PublishEvent(evt)

	h.logger.Info("Order status updated via HTTP", zap.String("order_id", orderID), zap.String("status", req.Status))
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Helper functions for the auth endpoints
func LoginHandler(authService *middleware.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Validate credentials against database
		// For now, create a token for demo purposes
		token, err := authService.GenerateToken("user123", req.Email, "user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":    "user123",
				"email": req.Email,
				"role":  "user",
			},
		})
	}
}

func RegisterHandler(authService *middleware.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Save user to database
		// For now, just return success
		token, err := authService.GenerateToken("user123", req.Email, "user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"token":   token,
			"user": gin.H{
				"id":    "user123",
				"email": req.Email,
				"role":  "user",
			},
		})
	}
}