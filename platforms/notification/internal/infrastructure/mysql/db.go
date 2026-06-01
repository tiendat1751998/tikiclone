package mysql
import ("fmt"; "github.com/jmoiron/sqlx"; _ "github.com/go-sql-driver/mysql"; "github.com/tikiclone/tiki/platforms/notification/internal/config")
func NewDB(cfg config.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil { return nil, fmt.Errorf("mysql connect: %w", err) }
	db.SetMaxOpenConns(cfg.MaxOpenConns); db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	if err := db.Ping(); err != nil { return nil, fmt.Errorf("mysql ping: %w", err) }
	db.MapperFunc(func(s string) string { return s }); return db, nil
}
