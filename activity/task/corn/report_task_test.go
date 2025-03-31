package cron_task

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestReplaceEnv(t *testing.T) {
	reportJsonDtoList := []*dto.ReportJsonDto{
		{Date: "111"},
		{Date: "222"},
	}
	ginCtx := gin.Context{}
	bytes, _ := generateExcelFile(&ginCtx, reportJsonDtoList)
	// 写入字节数组到文件
	err := ioutil.WriteFile("output.xlsx", bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("File written successfully!")
}

func TestEmail(t *testing.T) {
	emailConfig := config.ApplicationConfig.EmailConfig
	emailConfig.ServerHost = "smtp.163.com"
	emailConfig.ServerPort = 465
	emailConfig.FromAddress = ""
	emailConfig.ToAddressList = []string{""}
	//ginCtx := gin.Context{}
	//sendEmail(&ginCtx, []byte{"你好啊1111"})

	log.Println("File written successfully!")
}

func TestEmail22(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("第一个defer捕获到异常:", e)
			return
		}
	}()
	defer func() {

		// 这个defer中的recover不会执行，因为前面已经发生了panic
		if e := recover(); e != nil {
			fmt.Println("第二个defer捕获到异常:", e)
			panic(e)
		}
	}()

	panic("模拟一个异常")
	fmt.Println("正常执行的代码部分")
}

func TestTime(t *testing.T) {
	now := time.Now()
	customTime := util.NewCustomTime(now)
	unix := customTime.Unix()

	newNow := time.Unix(unix, 0)
	newCustomTime := util.NewCustomTime(newNow)
	log.Println(newCustomTime)
}

func TestLog(t *testing.T) {
	logTracing.LogInfo(context.Background(), "aaa")
	logTracing.LogWarn(context.Background(), logTracing.WarnLogFmt, "aaa")
	logTracing.LogError(context.Background(), errors.New("11"), "aaa")
}
