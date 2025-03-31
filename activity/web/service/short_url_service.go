package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/third/http_client"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util/config/encoder/json"
	"time"
)

type ShortUrlService struct {
}

func ShortUrlSign(value string, signKey string) string {
	hmac := hmac.New(sha1.New, []byte(signKey))
	hmac.Write([]byte(value))
	out := hmac.Sum(nil)
	return base64.StdEncoding.EncodeToString(out)

}

func getSignParamValue(longUrl string, activityId string, project string) string {
	return fmt.Sprintf("url=%s&expire_at=%d&activity_id=%s&project_id=%s", longUrl, 0, activityId, project)
}

func (u ShortUrlService) GetShortUrlByUrl(ctx *gin.Context, url, waId string) (string, error) {
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始调用生成短链接方法,waId:%v,url:%v", waId, url))
	shortDto := dto.ShortDto{
		LongUrl:         url,
		ActivityId:      "mcgg2025wa",
		ProjectId:       config.ApplicationConfig.Wa.McggShortProject,
		ShortLinkGenUrl: config.ApplicationConfig.Wa.McggShortLinkGenUrl,
		ShortLinkPrefix: config.ApplicationConfig.Wa.McggShortLinkPrefix,
		SignKey:         config.ApplicationConfig.Wa.McggShortLinkSignKey,
	}
	shortUrl, err := u.GetShortUrl(ctx, shortDto)

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("完成调用生成短链接方法,waId:%v,url:%v", waId, url))
	return shortUrl, err
}

func (u ShortUrlService) GetShortUrl(ctx *gin.Context, shortDto dto.ShortDto) (string, error) {
	methodName := "GetShortUrl"
	value := getSignParamValue(shortDto.LongUrl, shortDto.ActivityId, shortDto.ProjectId)
	sign := ShortUrlSign(value, shortDto.SignKey)

	params := map[string]any{
		"long_url":    shortDto.LongUrl,
		"expire_at":   0,
		"activity_id": shortDto.ActivityId,
		"project_id":  shortDto.ProjectId,
		"sign":        sign,
	}

	logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，开始调用生成短链接接口,请求：params:%v", methodName, params))

	res, nxErr := http_client.DoPostSSL(shortDto.ShortLinkGenUrl, params, nil, 10*1000*time.Second, 10*1000*time.Second)
	if nxErr != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，调用生成短链接接口http失败,params:%v,err:%v", methodName, params, nxErr))
		return "", nxErr
	}
	logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，结束调用生成短链接接口,请求：params:%v,返回: %v", methodName, params, res))

	resNx := &response.ShortLinkResponse{}
	nxErr = json.NewEncoder().Decode([]byte(res), resNx)
	if nxErr != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，生成短链接接口返回转实体报错,res:%v,err：%v", methodName, res, nxErr))
		return "", nxErr
	}
	if 0 != resNx.Code {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，调用生成短链接接口http失败,params:%v,res:%v", methodName, params, resNx))
		return "", errors.New("生成短链接失败")
	}
	return shortDto.ShortLinkPrefix + resNx.Data.ShortUrl, nil
}
