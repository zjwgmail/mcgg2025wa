package core

import (
	"context"
	"go-fission-activity/activity/web/middleware/logTracing"
	"net/http"
	_ "net/http/pprof" // 会自动注册 handler 到 http server，方便通过 http 接口获取程序运行采样报告
)

var srv http.Server

//
//func OpenPprof() error {
//	if config.ApplicationConfig.Pprof.Enable {
//		go startPprof()
//	}
//	currentPprofEnable := config.ApplicationConfig.Pprof.Enable
//	for {
//		if currentPprofEnable != config.ApplicationConfig.Pprof.Enable {
//			if config.ApplicationConfig.Pprof.Enable {
//				go startPprof()
//			} else {
//				go stopPprof()
//			}
//			currentPprofEnable = config.ApplicationConfig.Pprof.Enable
//		}
//		time.Sleep(time.Second * time.Duration(config.ApplicationConfig.Pprof.ListenEnableSecond))
//	}
//	return nil
//}

//func startPprof() {
//	// 开启对阻塞操作的跟踪
//	runtime.SetBlockProfileRate(1)
//	logTracing.LogPrintfP("启动 pprof")
//	// 启动一个 http server，注意 pprof 相关的 handler 已经自动注册过了
//	srv = initPprof()
//	err := srv.ListenAndServe()
//	if err != nil {
//		logTracing.LogPrintfP("pprof启动失败！error:%v \n", err.Error())
//	}
//}

func stopPprof() {
	logTracing.LogPrintfP("停止 pprof")
	err := srv.Shutdown(context.Background())
	if err != nil {
		logTracing.LogPrintfP("pprof停止失败！error:%v \n", err.Error())
	}

}

//func initPprof() http.Server {
//	return http.Server{Addr: fmt.Sprintf(":%v", config.ApplicationConfig.Pprof.Port), Handler: nil}
//}
