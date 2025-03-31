package middleware

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/middleware/logTracing"
)

// Trace 链路追踪
func Trace() gin.HandlerFunc {
	return logTracing.GinTracing
}
