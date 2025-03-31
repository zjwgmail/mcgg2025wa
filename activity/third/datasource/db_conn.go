package datasource

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/config"
	"sync"
	"time"
)

var engine GoMybatis.GoMybatisEngine
var goMybatisEngineOnce sync.Once

func GetConn() GoMybatis.GoMybatisEngine {
	goMybatisEngineOnce.Do(func() {
		engine = GoMybatis.GoMybatisEngine{}.New()
		//设置打印自动生成的xml 到控制台方便调试，false禁用
		engine.TemplateDecoder().SetPrintElement(true)
		//设置是否打印警告(建议开启)
		engine.SetPrintWarning(true)
		datasource := config.ApplicationConfig.Datasource
		db, err := engine.Open(datasource.DriverName, datasource.DataSourceLink)
		if err != nil {
			panic(err.Error())
		}
		db.SetMaxIdleConns(60)
		db.SetMaxOpenConns(60)
		db.SetConnMaxIdleTime(time.Second * datasource.MaxIdleTime)
		db.SetConnMaxLifetime(time.Second * datasource.MaxLifetime)

		if datasource.LogEnable { //打印SQL日志
			engine.SetLogEnable(true)
		}
		//conn.Engine.SetLog(&GoMybatis.LogStandard{
		//  PrintlnFunc: func(messages []byte) {
		//    fmt.Printf(messages)
		//  },
		//})
	})
	return engine
}
