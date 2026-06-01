// MongoDB indexes for 20M product scale
// Run: mongosh <mongodb-connection> tiki_catalog database/mongodb-indexes.js

db.products.createIndex({status: 1, 'skus.0.price': 1}, {name: 'idx_status_price_asc'});
db.products.createIndex({status: 1, 'skus.0.price': -1}, {name: 'idx_status_price_desc'});
db.products.createIndex({status: 1, sold_count: -1}, {name: 'idx_status_popularity'});
db.products.createIndex({status: 1, created_at: -1}, {name: 'idx_status_newest'});
db.products.createIndex({status: 1, category_id: 1, created_at: -1}, {name: 'idx_status_cat_newest'});
db.products.createIndex({status: 1, category_id: 1, 'skus.0.price': 1}, {name: 'idx_status_cat_price'});
db.products.createIndex({title: 'text', description: 'text'}, {name: 'idx_text_search', weights: {title: 10, description: 3}});

db.categories.createIndex({slug: 1}, {name: 'idx_cat_slug', unique: true});
db.categories.createIndex({category_id: 1}, {name: 'idx_cat_id', unique: true});

printjson(db.products.getIndexes().map(i => ({name: i.name, key: i.key})));
