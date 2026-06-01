package inventorypublic

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/inventory/internal/application"
	"github.com/tikiclone/tiki/services/inventory/internal/config"
	"github.com/tikiclone/tiki/services/inventory/internal/domain"
	"github.com/tikiclone/tiki/services/inventory/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/inventory/internal/infrastructure/redis"
)

type InventoryService = application.InventoryService

type ReserveStockRequest = application.ReserveStockRequest

type Reservation = domain.Reservation

type Stock = domain.Stock

type ReservationStatus = domain.ReservationStatus

const ReservationStatusActive = domain.ReservationStatusActive

type Config = config.Config

type InventoryConfig = config.InventoryConfig

type RedisConfig = config.RedisConfig

func NewInventoryRepository(db *sqlx.DB) *mysql.InventoryRepository {
	return mysql.NewInventoryRepository(db)
}

func NewRedisStore(client *redis.Client, cfg config.RedisConfig) *redisinfra.Store {
	return redisinfra.NewStore(client, cfg)
}

func NewInventoryService(cfg *config.Config, db *sql.DB, repo *mysql.InventoryRepository, store *redisinfra.Store) *InventoryService {
	return application.NewInventoryService(cfg, db, repo, store, nil)
}
