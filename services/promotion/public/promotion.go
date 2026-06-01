package promotionpublic

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/promotion/internal/application"
	"github.com/tikiclone/tiki/services/promotion/internal/config"
	"github.com/tikiclone/tiki/services/promotion/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/promotion/internal/infrastructure/redis"
)

type PromotionService = application.PromotionService

type RedisConfig = config.RedisConfig

func NewVoucherRepository(db *sqlx.DB) *mysql.VoucherRepository {
	return mysql.NewVoucherRepository(db)
}

func NewVoucherRedemptionRepository(db *sqlx.DB) *mysql.VoucherRedemptionRepository {
	return mysql.NewVoucherRedemptionRepository(db)
}

func NewRedisStore(client *redis.Client, cfg config.RedisConfig) *redisinfra.Store {
	return redisinfra.NewStore(client, cfg)
}

func NewPromotionService(voucherRepo *mysql.VoucherRepository, redemptionRepo *mysql.VoucherRedemptionRepository, redisStore *redisinfra.Store) *PromotionService {
	return application.NewPromotionService(voucherRepo, redemptionRepo, nil, nil, nil, nil, redisStore, nil)
}
