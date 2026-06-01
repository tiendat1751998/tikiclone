package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func NewMongoClient(uri, database string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx,
		options.Client().
			ApplyURI(uri).
			SetMinPoolSize(100).
			SetMaxPoolSize(1000).
			SetMaxConnIdleTime(30*time.Second).
			SetConnectTimeout(2*time.Second).
			SetSocketTimeout(2*time.Second).
			SetMonitor(otelmongo.NewMonitor()),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return nil, err
	}

	ensureIndexes(ctx, client.Database(database))

	return client, nil
}

func ensureIndexes(ctx context.Context, db *mongo.Database) {
	// Compound indexes for sort + filter — critical for 20M product scale
	priceAsc := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "skus.0.price", Value: 1}}, Options: options.Index().SetName("idx_status_price_asc")}
	priceDesc := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "skus.0.price", Value: -1}}, Options: options.Index().SetName("idx_status_price_desc")}
	popularity := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "sold_count", Value: -1}}, Options: options.Index().SetName("idx_status_popularity")}
	newest := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("idx_status_newest")}
	catNewest := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "category_id", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("idx_status_cat_newest")}
	catPrice := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "category_id", Value: 1}, {Key: "skus.0.price", Value: 1}}, Options: options.Index().SetName("idx_status_cat_price")}
	catPriceDesc := mongo.IndexModel{Keys: bson.D{{Key: "status", Value: 1}, {Key: "category_id", Value: 1}, {Key: "skus.0.price", Value: -1}}, Options: options.Index().SetName("idx_status_cat_price_desc")}
	textSearch := mongo.IndexModel{Keys: bson.D{{Key: "title", Value: "text"}, {Key: "description", Value: "text"}}, Options: options.Index().SetName("idx_text_search").SetWeights(bson.M{"title": 10, "description": 3})}
	spuID := mongo.IndexModel{Keys: bson.D{{Key: "spu_id", Value: 1}}, Options: options.Index().SetName("idx_spu_id").SetUnique(true)}

	db.Collection("products").Indexes().CreateMany(ctx, []mongo.IndexModel{priceAsc, priceDesc, popularity, newest, catNewest, catPrice, catPriceDesc, textSearch, spuID})

	// Category indexes
	catSlug := mongo.IndexModel{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetName("idx_cat_slug").SetUnique(true)}
	catID := mongo.IndexModel{Keys: bson.D{{Key: "category_id", Value: 1}}, Options: options.Index().SetName("idx_cat_id").SetUnique(true)}
	db.Collection("categories").Indexes().CreateMany(ctx, []mongo.IndexModel{catSlug, catID})
}
