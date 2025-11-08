package repository

import (
	"database/sql"
	"time"

	commonproto "food-delivery-platform/common/proto"
	"go.uber.org/zap"
)

type TrackingRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewTrackingRepo(db *sql.DB, logger *zap.Logger) *TrackingRepo {
	return &TrackingRepo{db: db, logger: logger}
}

func (r *TrackingRepo) UpdateLocation(deliveryId string, location *commonproto.Location) error {
	_, err := r.db.Exec(`
		INSERT INTO tracking (delivery_id, lat, lng, timestamp) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (delivery_id) 
		DO UPDATE SET lat = $2, lng = $3, timestamp = $4`,
		deliveryId, location.Lat, location.Lng, time.Now())
	if err != nil {
		r.logger.Error("Update location error", zap.Error(err))
		return err
	}
	return nil
}

func (r *TrackingRepo) GetLatestLocation(deliveryId string) (*commonproto.Location, error) {
	var location commonproto.Location
	var timestamp time.Time

	err := r.db.QueryRow(`
		SELECT lat, lng, timestamp 
		FROM tracking 
		WHERE delivery_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`, deliveryId).Scan(&location.Lat, &location.Lng, &timestamp)

	if err != nil {
		r.logger.Error("Get location error", zap.Error(err))
		return nil, err
	}

	location.Timestamp = timestamp.Format(time.RFC3339)
	return &location, nil
}

func (r *TrackingRepo) GetLocationHistory(deliveryId string, limit int) ([]*commonproto.Location, error) {
	rows, err := r.db.Query(`
		SELECT lat, lng, timestamp 
		FROM tracking 
		WHERE delivery_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2`, deliveryId, limit)

	if err != nil {
		r.logger.Error("Get location history error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var locations []*commonproto.Location
	for rows.Next() {
		var location commonproto.Location
		var timestamp time.Time

		err := rows.Scan(&location.Lat, &location.Lng, &timestamp)
		if err != nil {
			r.logger.Error("Scan location error", zap.Error(err))
			return nil, err
		}

		location.Timestamp = timestamp.Format(time.RFC3339)
		locations = append(locations, &location)
	}

	return locations, nil
}

func (r *TrackingRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS tracking (
		id SERIAL PRIMARY KEY,
		delivery_id VARCHAR(255),
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(delivery_id)
	)`)
	if err != nil {
		r.logger.Error("Init table error", zap.Error(err))
		return err
	}
	r.logger.Info("Tracking table initialized")
	return nil
}
