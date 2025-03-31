package middleware

import (
	"bufio"
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/util"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// LoggerToFile 日志记录到文件
func LoggerToFile() gin.HandlerFunc {

	return func(c *gin.Context) {
		// 开始时间
		startTime := util.GetNowCustomTime().Time
		// 处理请求
		var body string
		switch c.Request.Method {
		case http.MethodPost, http.MethodGet, http.MethodPut, http.MethodDelete:
			bf := bytes.NewBuffer(nil)
			wt := bufio.NewWriter(bf)
			_, err := io.Copy(wt, c.Request.Body)
			if err != nil {
				logTracing.LogPrintfP("copy body error, %s", err.Error())
				err = nil
			}
			rb, _ := ioutil.ReadAll(bf)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			body = string(rb)
		}

		c.Set("bodyStr", body)
		// 请求路由
		reqUri := c.Request.RequestURI

		c.Set("startTime", time.Now())
		c.Set("traceId", util.GetSnowFlakeIdStr(context.Background()))
		logTracing.LogPrintf(c, "[prod-mcgg2025wa-server]接收[web]请求:url:%s,body:%s", reqUri, body)

		c.Next()

		url := c.Request.RequestURI
		if strings.Index(url, "login") > -1 {
			return
		}
		//结束时间
		endTime := util.GetNowCustomTime().Time
		if c.Request.Method == http.MethodOptions {
			return
		}

		// 请求方式
		reqMethod := c.Request.Method

		// 状态码
		statusCode := c.Writer.Status()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 日志格式
		logTracing.LogPrintf(c, "[prod-mcgg2025wa-server]返回[web]结果:uri: %s,statusCode:%d, latencyTime: %s, method:%s", reqUri, statusCode, latencyTime, reqMethod)

	}
}
