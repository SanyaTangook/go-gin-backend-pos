package database

import (
	"fmt"

	"github.com/example/pos/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func Migrate(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		stock INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		total_amount DECIMAL(10, 2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transaction_items (
		id SERIAL PRIMARY KEY,
		transaction_id INT NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
		product_id INT NOT NULL REFERENCES products(id),
		product_name VARCHAR(255) NOT NULL,
		unit_price DECIMAL(10, 2) NOT NULL,
		quantity INT NOT NULL,
		subtotal DECIMAL(10, 2) NOT NULL
	);
	`

	_, err := db.Exec(schema)
	return err
}
