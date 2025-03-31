package router

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/config"
	"log"
	"os"
)

// 路由初始化
func InitRouter() {

	var r *gin.Engine
	h := config.Runtime.GetEngine()
	if h == nil {
		log.Fatal("not found engine...")
		os.Exit(-1)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
		os.Exit(-1)
	}

	//注册业务路由
	initRouter(r)
}
