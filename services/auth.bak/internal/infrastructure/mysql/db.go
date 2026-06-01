package mysql

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tikiclone/tiki/services/auth/internal/config"
)

func NewDB(cfg config.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("mysql connect: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping: %w", err)
	}

	db.MapperFunc(func(s string) string { return s })

	return db, nil
}

func NewTestDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn+"?parseTime=true&loc=UTC")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.MapperFunc(func(s string) string { return s })
	return db, nil
}
