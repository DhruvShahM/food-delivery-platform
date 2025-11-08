package repository

import (
	"database/sql"

	"go.uber.org/zap"
)

type PaymentRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPaymentRepo(db *sql.DB, logger *zap.Logger) *PaymentRepo {
	return &PaymentRepo{db: db, logger: logger}
}

func (r *PaymentRepo) Process(userId string, amount float64) error {
	_, err := r.db.Exec("INSERT INTO payments (user_id, amount, status) VALUES ($1, $2, 'completed')", userId, amount)
	if err != nil {
		r.logger.Error("Process payment error", zap.Error(err))
		return err
	}
	return nil
}

func (r *PaymentRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255),
		amount DOUBLE PRECISION,
		status VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		r.logger.Error("Init table error", zap.Error(err))
		return err
	}
	r.logger.Info("Payments table initialized")
	return nil
}

func (r *PaymentRepo) GetBalance(userId string) (float64, error) {
	var balance float64
	err := r.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE user_id = $1 AND status = 'completed'", userId).Scan(&balance)
	if err != nil {
		r.logger.Error("Get balance error", zap.Error(err))
		return 0, err
	}
	return balance, nil
}

func (r *PaymentRepo) Refund(transactionId string, amount float64) error {
	// For simplicity, we'll insert a negative payment record
	// In a real system, you'd update the original transaction
	userId := "user123" // TODO: Get from transaction lookup

	_, err := r.db.Exec("INSERT INTO payments (user_id, amount, status, transaction_id) VALUES ($1, $2, 'refunded', $3)",
		userId, -amount, transactionId)
	if err != nil {
		r.logger.Error("Refund error", zap.Error(err))
		return err
	}
	return nil
}
