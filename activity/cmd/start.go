package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/core"
	cron_task "go-fission-activity/activity/task/corn"
	"go-fission-activity/activity/third/datasource"
	"go-fission-activity/activity/third/http_client"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/router"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/file"
	"log"
	"os"
)

var (
	configYml    string
	msgConfigYml string
	apiCheck     bool
	StartCmd     = &cobra.Command{
		Use:          "start",
		Short:        "prod-mcgg2025wa-server",
		Example:      constant.AppName + " start -c " + constant.AppYaml,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", constant.AppYaml, "使用提供的配置文件启动服务器")
	StartCmd.PersistentFlags().StringVarP(&msgConfigYml, "msgConfig", "m", constant.MsgConfigYaml, "使用提供的消息配置文件启动服务器")
	StartCmd.PersistentFlags().BoolVarP(&apiCheck, "api", "a", false, "使用检查api数据启动服务器")

	//初始化路由
	core.AppRouters = append(core.AppRouters, router.InitRouter)
}

func run() error {
	//go core.OpenPprof()
	return core.RunWeb()
}

func setup() {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(fmt.Sprintf("无法获取当前工作目录：%v", err))
		return
	}
	log.Println("当前工作目录：" + currentDir)
	log.Printf("当前配置文件环境：" + constant.ProfileActives)
	log.Println("初始化配置文件:" + configYml)
	log.Println("初始化消息配置文件:" + msgConfigYml)
	// 注入配置扩展项
	config.ExtendConfig = &config.ExtConfig
	//1. 读取配置
	config.Setup(file.NewSource(file.WithPath(configYml)))

	scheme := config.ApplicationConfig.Activity.Scheme
	log.Printf("使用消息方案：%v", scheme)
	log.Printf("使用消息配置文件：%v", msgConfigYml)
	config.MsgSetup(file.NewSource(file.WithPath(msgConfigYml)))
	usageStr := `starting api server...`
	log.Println(usageStr)
	log.Printf("配置文件内容:%+v", config.ApplicationConfig)
	log.Printf("消息配置文件内容:%+v", config.MsgConfig)
	//初始化数据库连接
	datasource.MybatisInit()
	// 初始化redis连接
	redis_template.NewRedisTemplate()
	service.InitHelpWeight(context.Background())
	service.InitImageService()
	http_client.InitHttpClientPool()

	util.GetLocation()
	// 初始化sentry
	go core.OpenZap()

	go cron_task.InitCron()
}
