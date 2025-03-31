package datasource

import (
	"go-fission-activity/activity/web/middleware/logTracing"
	"log"
	"os"
)

func MybatisInit() {
	logTracing.LogPrintfP("开始初始化mybatis配置文件，初始化数据库连接")
	GetConn()
	MybatisMapperInit()
}

func InitMybatisXMLConfig(xml string, ptr interface{}) {
	conn := GetConn()
	bytes, err := os.ReadFile(xml)
	if nil != err {
		log.Fatalf("加载%s失败%v", xml, err)
	}
	conn.WriteMapperPtr(ptr, bytes)
}
