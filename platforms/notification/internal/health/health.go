package health

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
)

type Checker struct {
	svc string
	ver string
}

func NewChecker(svc, version string) *Checker {
	return &Checker{svc: svc, ver: version}
}

func (c *Checker) LivenessHandler() gin.HandlerFunc {
	hc := health.NewChecker(c.svc, c.ver)
	return hc.LivenessHandler()
}

func (c *Checker) ReadinessHandler() gin.HandlerFunc {
	hc := health.NewChecker(c.svc, c.ver)
	return hc.ReadinessHandler()
}
