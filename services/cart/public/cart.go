package cartpublic

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/cart/internal/application"
	"github.com/tikiclone/tiki/services/cart/internal/infrastructure/mysql"
)

type CartService = application.CartService

type AddItemRequest = application.AddItemRequest

func NewCartRepository(db *sqlx.DB) *mysql.CartRepository {
	return mysql.NewCartRepository(db)
}

func NewCartItemRepository(db *sqlx.DB) *mysql.CartItemRepository {
	return mysql.NewCartItemRepository(db)
}

func NewCartService(
	cartRepo *mysql.CartRepository,
	itemRepo *mysql.CartItemRepository,
	cartTTL time.Duration,
	checkoutPreviewTTL time.Duration,
	maxCartItems int,
	maxQuantity int,
) *CartService {
	return application.NewCartService(
		cartRepo, itemRepo, nil, nil,
		nil,
		cartTTL, checkoutPreviewTTL, maxCartItems, maxQuantity, nil,
	)
}
