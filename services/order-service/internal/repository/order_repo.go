package repository

import (
	"database/sql"

	"food-delivery-platform/services/order-service/internal/proto"
	commonproto "food-delivery-platform/common/proto"
	"food-delivery-platform/common/pkg/kafka"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type OrderRepo struct {
	db       *sql.DB
	producer *kafkago.Writer
	logger   *zap.Logger
}

func NewOrderRepo(db *sql.DB, producer *kafkago.Writer, logger *zap.Logger) *OrderRepo {
	return &OrderRepo{db: db, producer: producer, logger: logger}
}

func (r *OrderRepo) CreateOrder(orderId, userId, restaurantId string, amount float64) error {
	_, err := r.db.Exec("INSERT INTO orders (id, user_id, restaurant_id, amount, status) VALUES ($1, $2, $3, $4, 'created')", orderId, userId, restaurantId, amount)
	if err != nil {
		r.logger.Error("Create order error", zap.Error(err))
		return err
	}
	return nil
}

func (r *OrderRepo) PublishEvent(evt *commonproto.OrderEvent) error {
	return kafka.Publish(r.producer, evt, r.logger)
}

func (r *OrderRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255),
		restaurant_id VARCHAR(255),
		amount DOUBLE PRECISION,
		status VARCHAR(50)
	)`)
	if err != nil {
		r.logger.Error("Init table error", zap.Error(err))
		return err
	}
	r.logger.Info("Orders table initialized")
	return nil
}

func (r *OrderRepo) GetOrdersByUserID(userID string) ([]*proto.Order, error) {
	rows, err := r.db.Query("SELECT id, user_id, restaurant_id, amount, status FROM orders WHERE user_id = $1", userID)
	if err != nil {
		r.logger.Error("Get orders error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []*proto.Order
	for rows.Next() {
		var order proto.Order
		err := rows.Scan(&order.Id, &order.UserId, &order.RestaurantId, &order.Amount, &order.Status)
		if err != nil {
			r.logger.Error("Scan order error", zap.Error(err))
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (r *OrderRepo) UpdateOrderStatus(orderID, status string) error {
	_, err := r.db.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	if err != nil {
		r.logger.Error("Update order status error", zap.Error(err))
		return err
	}
	return nil
}