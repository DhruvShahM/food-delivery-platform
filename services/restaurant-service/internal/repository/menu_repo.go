package repository

import (
	"database/sql"

	commonproto "food-delivery-platform/common/proto"
	"go.uber.org/zap"
)

type MenuRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewMenuRepo(db *sql.DB, logger *zap.Logger) *MenuRepo {
	return &MenuRepo{db: db, logger: logger}
}

func (r *MenuRepo) GetMenu(restaurantId string) ([]*commonproto.MenuItem, error) {
	rows, err := r.db.Query("SELECT id, name, price, available FROM menu WHERE restaurant_id = $1", restaurantId)
	if err != nil {
		r.logger.Error("Query menu error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []*commonproto.MenuItem
	for rows.Next() {
		item := &commonproto.MenuItem{}
		err := rows.Scan(&item.Id, &item.Name, &item.Price, &item.Available)
		if err != nil {
			r.logger.Error("Scan error", zap.Error(err))
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MenuRepo) UpdateAvailability(itemId string, available bool) error {
	_, err := r.db.Exec("UPDATE menu SET available = $1 WHERE id = $2", available, itemId)
	if err != nil {
		r.logger.Error("Update error", zap.Error(err))
		return err
	}
	return nil
}

func (r *MenuRepo) CreateMenuItem(restaurantId, id, name string, price float64) error {
	_, err := r.db.Exec("INSERT INTO menu (id, restaurant_id, name, price, available) VALUES ($1, $2, $3, $4, true)", id, restaurantId, name, price)
	if err != nil {
		r.logger.Error("Create error", zap.Error(err))
		return err
	}
	return nil
}

func (r *MenuRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS menu (
		id VARCHAR(255) PRIMARY KEY,
		restaurant_id VARCHAR(255),
		name VARCHAR(255),
		price DOUBLE PRECISION,
		available BOOLEAN
	)`)
	if err != nil {
		r.logger.Error("Init table error", zap.Error(err))
		return err
	}
	// Sample data
	_, err = r.db.Exec(`INSERT INTO menu (id, restaurant_id, name, price, available) VALUES 
		('1', 'rest1', 'Pizza', 10.0, true),
		('2', 'rest1', 'Burger', 5.0, true) 
		ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		r.logger.Warn("Sample data insert", zap.Error(err))
	}
	r.logger.Info("Menu table initialized")
	return nil
}