package config

import (
	"fmt"
	"go-fission-activity/activity/constant"
	"go-fission-activity/util/config"
	"go-fission-activity/util/config/source"
	"log"
	"os"
)

var (
	ExtendConfig interface{}
	_cfg         *Settings
)

// Settings 兼容原先的配置结构
type Settings struct {
	Settings  Config `yaml:"settings"`
	callbacks []func()
}

func (e *Settings) runCallback() {
	for i := range e.callbacks {
		e.callbacks[i]()
	}
}

func (e *Settings) OnChange() {
	e.init()
	log.Println("!!! config change and reload")
}

func (e *Settings) Init() {
	e.init()
	log.Println("!!! config init")
}

func (e *Settings) init() {
	log.Println("change init..........")
	e.runCallback()
}

// Config 配置集合
type Config struct {
	Application *Application `yaml:"application"`
	Extend      interface{}  `yaml:"extend"`
}

// Setup 载入配置文件
func Setup(s source.Source, fs ...func()) {
	_cfg = &Settings{
		Settings: Config{
			Application: ApplicationConfig,
			Extend:      ExtendConfig,
		},
		callbacks: fs,
	}
	var err error
	config.DefaultConfig, err = config.NewConfig(
		config.WithSource(s),
		config.WithEntity(_cfg),
	)
	if err != nil {
		log.Fatal(fmt.Sprintf("New config object fail: %s", err.Error()))
	}
	_cfg.Init()
	replaceConfigByEnv()
}

func replaceConfigByEnv() {
	env := constant.LookupEnv(constant.ProfileActives, constant.DefaultActives)

	// 如果不是本地环境，就读取环境变量
	if constant.DefaultActives != env {

		S3PreSignUrl := os.Getenv("S3_PRE_SIGNED_URL")
		if S3PreSignUrl == constant.Empty {
			log.Fatalln("S3_PRE_SIGNED_URL 变量不能为空")
		} else {
			ApplicationConfig.S3Config.PreSignUrl = S3PreSignUrl
			log.Printf("S3_PRE_SIGNED_URL 更新为：%v", ApplicationConfig.S3Config.PreSignUrl)
		}

		//DonAmin := os.Getenv("DonAmin")
		//if DonAmin == constant.Empty {
		//	log.Fatalln("DonAmin 变量不能为空")
		//} else {
		//	ApplicationConfig.S3Config.DonAmin = DonAmin
		//	log.Printf("DonAmin 更新为：%v", ApplicationConfig.S3Config.PreSignUrl)
		//}

		Ak := os.Getenv("NX_AK")
		if Ak == constant.Empty {
			log.Fatalln("NX_AK 变量不能为空")
		} else {
			ApplicationConfig.Nx.Ak = Ak
			log.Printf("NX_AK 更新为：%v", ApplicationConfig.Nx.Ak)
		}

		Sk := os.Getenv("NX_SK")
		if Sk == constant.Empty {
			log.Fatalln("NX_SK 变量不能为空")
		} else {
			ApplicationConfig.Nx.Sk = Sk
			log.Printf("NX_SK 更新为：%v", ApplicationConfig.Nx.Sk)
		}

		RedisAddress := os.Getenv("REDIS_ADDRESS")
		if RedisAddress == constant.Empty {
			log.Fatalln("REDIS_ADDRESS 变量不能为空")
		} else {
			ApplicationConfig.Redis.Address = RedisAddress
			log.Printf("REDIS_ADDRESS 更新为：%v", ApplicationConfig.Redis.Address)
		}

		RedisUsername := os.Getenv("REDIS_USERNAME")
		//if RedisUsername == constant.Empty {
		//	log.Fatalln("REDIS_USERNAME 变量不能为空")
		//} else {
		ApplicationConfig.Redis.Username = RedisUsername
		log.Printf("REDIS_USERNAME 更新为：%v", ApplicationConfig.Redis.Username)
		//}

		RedisPassword := os.Getenv("REDIS_PASSWORD")
		//if RedisPassword == constant.Empty {
		//	log.Fatalln("REDIS_PASSWORD 变量不能为空")
		//} else {
		ApplicationConfig.Redis.Password = RedisPassword
		log.Printf("REDIS_PASSWORD 更新为：%v", ApplicationConfig.Redis.Password)
		//}

		DataSourceLink := os.Getenv("DATA_SOURCE_LINK")
		if DataSourceLink == constant.Empty {
			log.Fatalln("DATA_SOURCE_LINK 变量不能为空")
		} else {
			ApplicationConfig.Datasource.DataSourceLink = DataSourceLink
			log.Printf("DATA_SOURCE_LINK 更新为：%v", ApplicationConfig.Datasource.DataSourceLink)
		}

	}
}
