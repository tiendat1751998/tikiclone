package health
import ("context"; "github.com/gin-gonic/gin"; "github.com/tikiclone/tiki/packages/go-shared/pkg/health"; "github.com/redis/go-redis/v9")
type Checker struct { redis *redis.Client; httpHealth *health.Checker }
func NewChecker(s, v string, r *redis.Client) *Checker {
	c := &Checker{redis: r, httpHealth: health.NewChecker(s, v)}
	if r != nil { c.httpHealth.AddCheck("redis", func(ctx context.Context) error { return r.Ping(ctx).Err() }) }
	return c
}
func (c *Checker) LivenessHandler() gin.HandlerFunc { return c.httpHealth.LivenessHandler() }
func (c *Checker) ReadinessHandler() gin.HandlerFunc { return c.httpHealth.ReadinessHandler() }
func (c *Checker) HealthChecker() *health.Checker { return c.httpHealth }
