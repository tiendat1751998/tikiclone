package health

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
)

type Checker struct {
	db         interface{ PingContext(ctx context.Context) error }
	redis      interface{ Ping(ctx context.Context) error }
	httpHealth *health.Checker
}

func NewChecker(service, version string, db interface{ PingContext(ctx context.Context) error }, rdb interface{ Ping(ctx context.Context) error }) *Checker {
	c := &Checker{db: db, redis: rdb, httpHealth: health.NewChecker(service, version)}
	c.httpHealth.AddCheck("database", func(ctx context.Context) error { return db.PingContext(ctx) })
	if rdb != nil {
		c.httpHealth.AddCheck("redis", func(ctx context.Context) error { return rdb.Ping(ctx) })
	}
	return c
}

func (c *Checker) LivenessHandler() gin.HandlerFunc  { return c.httpHealth.LivenessHandler() }
func (c *Checker) ReadinessHandler() gin.HandlerFunc { return c.httpHealth.ReadinessHandler() }
