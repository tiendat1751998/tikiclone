-- Migration: 004_fix_currency_sgd_to_vnd
-- Convert all SGD orders to VND (historical data correction)
-- VND is the only supported currency

UPDATE orders SET currency = 'VND' WHERE currency = 'SGD' OR currency = '' OR currency IS NULL;
UPDATE order_snapshots SET snapshot_data = JSON_SET(snapshot_data, '$.currency', 'VND') WHERE JSON_EXTRACT(snapshot_data, '$.currency') = 'SGD';
