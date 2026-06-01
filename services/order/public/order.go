package orderpublic

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/order/internal/application"
	"github.com/tikiclone/tiki/services/order/internal/config"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	"github.com/tikiclone/tiki/services/order/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/order/internal/infrastructure/redis"
)

type OrderService = application.OrderService

type CreateOrderRequest = application.CreateOrderRequest

type Order = domain.Order

type OrderStatus = domain.OrderStatus

type OrderItem = domain.OrderItem

type SnapshotItem = domain.SnapshotItem

const (
	OrderStatusPending         = domain.OrderStatusPending
	OrderStatusAwaitingPayment = domain.OrderStatusAwaitingPayment
	OrderStatusPaid            = domain.OrderStatusPaid
	OrderStatusProcessing      = domain.OrderStatusProcessing
	OrderStatusPacked          = domain.OrderStatusPacked
	OrderStatusShipped         = domain.OrderStatusShipped
	OrderStatusDelivered       = domain.OrderStatusDelivered
	OrderStatusCompleted       = domain.OrderStatusCompleted
	OrderStatusCancelled       = domain.OrderStatusCancelled
	OrderStatusRefunded        = domain.OrderStatusRefunded
)

type Config = config.Config

type OrderConfig = config.OrderConfig

type RedisConfig = config.RedisConfig

func NewOrderRepository(db *sqlx.DB) *mysql.OrderRepository {
	return mysql.NewOrderRepository(db)
}

func NewOutboxRepository(db *sqlx.DB) *mysql.OutboxRepository {
	return mysql.NewOutboxRepository(db)
}

func NewRedisStore(client *redis.Client, cfg config.RedisConfig) *redisinfra.Store {
	return redisinfra.NewStore(client, cfg)
}

func NewOrderService(cfg *config.Config, orderRepo *mysql.OrderRepository, outboxRepo *mysql.OutboxRepository, redisStore *redisinfra.Store) *OrderService {
	return application.NewOrderService(cfg, orderRepo, outboxRepo, redisStore, nil)
}
