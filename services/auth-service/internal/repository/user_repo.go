package repository

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewUserRepo(db *sql.DB, logger *zap.Logger) *UserRepo {
	return &UserRepo{db: db, logger: logger}
}

func (r *UserRepo) GetUserByEmail(email string) (*User, error) {
	row := r.db.QueryRow("SELECT id, email, password FROM users WHERE email = $1", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("User not found", zap.String("email", email))
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Query error", zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) CreateUser(email, password string) error {
	hashed := []byte(password) // Simple for demo, use bcrypt in prod
	_, err := r.db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, hashed)
	if err != nil {
		r.logger.Error("Create user error", zap.Error(err))
		return err
	}
	return nil
}

func (r *UserRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`)
	if err != nil {
		r.logger.Error("Init table error", zap.Error(err))
		return err
	}
	r.logger.Info("Users table initialized")
	return nil
}
