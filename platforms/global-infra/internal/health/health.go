package health

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
)

type Checker struct {
	httpHealth *health.Checker
}

func NewChecker(s, v string) *Checker {
	return &Checker{httpHealth: health.NewChecker(s, v)}
}

func (c *Checker) LivenessHandler() gin.HandlerFunc { return c.httpHealth.LivenessHandler() }
func (c *Checker) ReadinessHandler() gin.HandlerFunc { return c.httpHealth.ReadinessHandler() }
