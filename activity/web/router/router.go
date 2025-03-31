package router

import (
	"github.com/gin-gonic/gin"
)

var (
	RouterNoCheckRole = make([]func(*gin.RouterGroup), 0)
)

// initRouter 路由示例
func initRouter(r *gin.Engine) *gin.Engine {

	// 无需认证的路由
	noCheckRoleRouter(r)
	return r
}

// noCheckRoleRouter 无需认证的路由示例
func noCheckRoleRouter(r *gin.Engine) {
	// 可根据业务需求来设置接口版本
	v1 := r.Group("/")

	for _, f := range RouterNoCheckRole {
		f(v1)
	}
}
