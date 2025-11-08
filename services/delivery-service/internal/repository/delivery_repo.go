package repository

import (
	"database/sql"

	"food-delivery-platform/services/delivery-service/internal/proto"
	"go.uber.org/zap"
)

type DeliveryRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDeliveryRepo(db *sql.DB, logger *zap.Logger) *DeliveryRepo {
	return &DeliveryRepo{db: db, logger: logger}
}

func (r *DeliveryRepo) AssignDelivery(orderId, agentId string) error {
	_, err := r.db.Exec("INSERT INTO deliveries (order_id, agent_id, status) VALUES ($1, $2, 'assigned')", orderId, agentId)
	if err != nil {
		r.logger.Error("Assign error", zap.Error(err))
		return err
	}
	return nil
}

func (r *DeliveryRepo) Assign(orderId, agentId string) error {
	_, err := r.db.Exec("INSERT INTO deliveries (order_id, agent_id, status) VALUES ($1, $2, 'assigned')", orderId, agentId)
	if err != nil {
		r.logger.Error("Assign error", zap.Error(err))
		return err
	}
	return nil
}

func (r *DeliveryRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS deliveries (
		id SERIAL PRIMARY KEY,
		order_id VARCHAR(255),
		agent_id VARCHAR(255),
		status VARCHAR(50)
	)`)
	if err != nil {
		r.logger.Error("Init table error", zap.Error(err))
		return err
	}
	r.logger.Info("Deliveries table initialized")
	return nil
}

func (r *DeliveryRepo) GetDeliveriesByAgentID(agentID string) ([]*proto.Delivery, error) {
	rows, err := r.db.Query("SELECT id, order_id, agent_id, status FROM deliveries WHERE agent_id = $1", agentID)
	if err != nil {
		r.logger.Error("Get deliveries error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var deliveries []*proto.Delivery
	for rows.Next() {
		var delivery proto.Delivery
		err := rows.Scan(&delivery.Id, &delivery.OrderId, &delivery.AgentId, &delivery.Status)
		if err != nil {
			r.logger.Error("Scan delivery error", zap.Error(err))
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}
	return deliveries, nil
}

func (r *DeliveryRepo) UpdateDeliveryStatus(agentID, status string) error {
	_, err := r.db.Exec("UPDATE deliveries SET status = $1 WHERE agent_id = $2", status, agentID)
	if err != nil {
		r.logger.Error("Update delivery status error", zap.Error(err))
		return err
	}
	return nil
}
