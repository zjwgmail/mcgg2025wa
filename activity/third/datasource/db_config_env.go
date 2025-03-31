package datasource

import (
	"fmt"
	"go-fission-activity/activity/constant"
	"go-fission-activity/config"
	"log"
	"os"
	"strconv"
	"strings"
)

type DbConfigEnv struct {
}

func (e *DbConfigEnv) ReplaceEnv() {
	//host=172.16.100.159 port=5432 user=postgres password=postgres dbname=brainstorm_two
	/**
	DB_PASSWORD=t7zR0GTGStVCjqQuncMt
	DB_USERNAME=postgres
	DB_URL=jdbc:postgresql://172.16.100.229:5432
	*/
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == constant.Empty {
		log.Fatalln("DB_URL地址变量不能为空, 示例 jdbc:postgresql://172.16.100.229:5432")
		return
	}
	dbDbname := os.Getenv("DB_NAME")
	if dbDbname == constant.Empty {
		dbDbname = "brainstorm_two"
	}

	dbUserName := os.Getenv("DB_USERNAME")
	if dbUserName == constant.Empty {
		log.Fatalln("DB_USERNAME变量不能为空, 示例 root")
		return
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == constant.Empty {
		log.Fatalln("DB_PASSWORD变量不能为空, 示例 123456")
		return
	}
	var dbUrlPrefix = "jdbc:postgresql://"

	if strings.Contains(dbUrl, dbUrlPrefix) {
		dbSplitUrl := strings.Split(strings.Replace(dbUrl, "jdbc:postgresql://", "", 1), ":")
		config.ApplicationConfig.Datasource.DataSourceLink = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbSplitUrl[0], dbSplitUrl[1], dbUserName, dbPassword, dbDbname)
		config.ApplicationConfig.Datasource.DriverName = "postgres"
		log.Printf("Datasource.DataSourceLink 更新为：%v", config.ApplicationConfig.Datasource.DataSourceLink)
		log.Printf("Datasource.DriverName 更新为：%v", config.ApplicationConfig.Datasource.DriverName)
	} else {
		log.Fatalln("目前该服务只支持postgres数据库的解析操作")
		return
	}
	maxIdleCount := constant.LookupEnv("DB_MAX_IDLE_COUNT", "")
	if maxIdleCount != constant.Empty {
		atoi, err := strconv.Atoi(maxIdleCount)
		if nil == err {
			config.ApplicationConfig.Datasource.MaxIdleCount = atoi
			log.Printf("Datasource.MaxIdleCount 更新为：%v", atoi)
		}
	}

	maxOpen := constant.LookupEnv("DB_MAX_OPEN_COUNT", "")
	if maxOpen != constant.Empty {
		atoi, err := strconv.Atoi(maxOpen)
		if nil == err {
			config.ApplicationConfig.Datasource.MaxOpen = atoi
			log.Printf("Datasource.MaxOpen 更新为：%v", atoi)
		}
	}

	logEnable := constant.LookupEnv("DB_LOG_ENABLE", "")
	if logEnable != constant.Empty {
		parseBool, err := strconv.ParseBool(logEnable)
		if nil == err {
			config.ApplicationConfig.Datasource.LogEnable = parseBool
			log.Printf("Datasource.LogEnable 更新为：%s", logEnable)
		}
	}

}
