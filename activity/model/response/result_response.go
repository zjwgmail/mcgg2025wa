package response

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/middleware/logTracing"
)

type ResultResponse struct {

	//状态码
	Code int `json:"code"`

	//返回消息
	Message string `json:"message"`

	//返回消息
	TraceId string `json:"traceId"`

	// 响应信息
	Data interface{} `json:"data"`
}

func ResError(context *gin.Context, msg string) *ResultResponse {
	resultResponse := &ResultResponse{}
	resultResponse.Message = msg
	resultResponse.Code = 400
	_, traceId := logTracing.GetInfoByContext(context)
	resultResponse.TraceId = traceId
	context.JSON(200, resultResponse)
	return resultResponse
}

func ResSuccess(context *gin.Context, data interface{}) *ResultResponse {
	resultResponse := &ResultResponse{}
	resultResponse.Message = "操作成功"
	resultResponse.Code = 200
	resultResponse.Data = data
	_, traceId := logTracing.GetInfoByContext(context)
	resultResponse.TraceId = traceId
	context.JSON(200, resultResponse)
	return resultResponse
}
