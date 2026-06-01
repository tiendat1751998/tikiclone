package mysql

import (
	"github.com/jmoiron/sqlx"
)

func NewDB(cfg interface{ DSN() string }) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	return db, nil
}