module github.com/tikiclone/tiki/tests/integration

go 1.26.3

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/tikiclone/tiki/packages/go-shared v0.0.0-00010101000000-000000000000
	github.com/tikiclone/tiki/services/cart v0.0.0-00010101000000-000000000000
	github.com/tikiclone/tiki/services/inventory v0.0.0-00010101000000-000000000000
	github.com/tikiclone/tiki/services/order v0.0.0-00010101000000-000000000000
	github.com/tikiclone/tiki/services/payment v0.0.0-00010101000000-000000000000
	github.com/tikiclone/tiki/services/promotion v0.0.0-00010101000000-000000000000
)

replace (
	github.com/tikiclone/tiki/packages/go-shared => ../../packages/go-shared
	github.com/tikiclone/tiki/services/cart => ../../services/cart
	github.com/tikiclone/tiki/services/inventory => ../../services/inventory
	github.com/tikiclone/tiki/services/order => ../../services/order
	github.com/tikiclone/tiki/services/payment => ../../services/payment
	github.com/tikiclone/tiki/services/promotion => ../../services/promotion
)
