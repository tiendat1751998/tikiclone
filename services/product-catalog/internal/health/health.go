package health
import ("context"; "github.com/gin-gonic/gin"; "github.com/tikiclone/tiki/packages/go-shared/pkg/health"; "github.com/redis/go-redis/v9")
type Checker struct { db interface{ PingContext(ctx context.Context) error }; redis *redis.Client; httpHealth *health.Checker }
func NewChecker(s, v string, db interface{ PingContext(ctx context.Context) error }, r *redis.Client) *Checker {
	c := &Checker{db: db, redis: r, httpHealth: health.NewChecker(s, v)}
	c.httpHealth.AddCheck("database", func(ctx context.Context) error { return db.PingContext(ctx) })
	if r != nil { c.httpHealth.AddCheck("redis", func(ctx context.Context) error { return r.Ping(ctx).Err() }) }
	return c
}
func (c *Checker) LivenessHandler() gin.HandlerFunc { return c.httpHealth.LivenessHandler() }
func (c *Checker) ReadinessHandler() gin.HandlerFunc { return c.httpHealth.ReadinessHandler() }
