package core

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/web/middleware"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var AppRouters = make([]func(), 0)

func RunWeb() error {
	if constant.LookupEnv(constant.ProfileActives, constant.DefaultActives) == constant.ProdActives {
		gin.SetMode(gin.ReleaseMode)
	}
	initRouter()

	for _, f := range AppRouters {
		f()
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.ApplicationConfig.Host, config.ApplicationConfig.Port),
		Handler: config.Runtime.GetEngine(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen: ", err)
		}
	}()
	logTracing.LogPrintf(ctx, "web server start success! host:%s, port:%d", config.ApplicationConfig.Host, config.ApplicationConfig.Port)
	tip()
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Printf("%s Shutdown Server ... \r\n", time.Now().Format("2006-01-02 15:04:05"))

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Fatal("Server exiting")

	return nil
}

func initRouter() {
	var r *gin.Engine
	e := config.Runtime.GetEngine()
	if e == nil {
		e = gin.New()
		config.Runtime.SetEngine(e)
	}
	switch e.(type) {
	case *gin.Engine:
		r = e.(*gin.Engine)
	default:
		log.Fatal("不支持其它Engine")

	}

	middleware.InitMiddleware(r)

}

func tip() {
	usageStr := `欢迎使用 ` + config.ApplicationConfig.Name + ` 可以使用 -h 查看命令`
	fmt.Printf("%s \n\n", usageStr)
}
