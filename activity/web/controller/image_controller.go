package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/orderedmap"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/nx"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"
)

type ImageController struct {
	ImageService    *service.ImageService
	WaMsgService    *service.WaMsgService
	ShortUrlService *service.ShortUrlService
}

func (c ImageController) PreSign(ctx *gin.Context) {
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("PreSign，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	preSignParam := &request.PreSignParam{}
	ctx.ShouldBindJSON(preSignParam)

	res, err := c.ImageService.GeneratePreSignedURL(ctx, preSignParam)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[PreSign]，生成预签名URL失败,preSignParam:%v，err:%v", preSignParam, err))
		response.ResError(ctx, "generate preSign url error")
		return
	}
	response.ResSuccess(ctx, res)
}

func (c ImageController) UploadTemplateImage2NX(ctx *gin.Context) {
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("UploadTemplateImage2NX，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()
	path := "/Users/zhangjianwu/Documents/mlbb/fission/通知封面.png"
	nx1, err := c.WaMsgService.UploadTemplateImage2NX(ctx, path)
	logTracing.LogInfo(ctx, fmt.Sprintf("nx1:%v", nx1))
	nx2, err := c.WaMsgService.UploadMedia2NX(ctx, path)
	logTracing.LogInfo(ctx, fmt.Sprintf("nx2:%v", nx2))
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[UploadTemplateImage2NX]，上传图片到NX失败,err:%v", err))
		response.ResError(ctx, "upload image to NX error")
		return
	}
	response.ResSuccess(ctx, nx2)
}

func (c ImageController) GenerateImages(ctx *gin.Context) {
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("GenerateImages，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	param := &request.SynthesisParam{}
	ctx.ShouldBindJSON(param)
	bizType := param.BizType
	if bizType == 1 {
		res, err := c.ImageService.GetInteractiveImageUrl(ctx, param, "")
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GenerateImages]，生成互动图片失败,param:%v，err:%v", param, err))
			response.ResError(ctx, "generate interactive image error")
			return
		}
		response.ResSuccess(ctx, res)
	} else if bizType == 2 {
		path := param.FilePath
		imageData, err := ioutil.ReadFile(path)
		res, err := c.ImageService.GenerateImageAndUpload2S3WithBytes(imageData, ctx, param, "")
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GenerateImages]，生成互动图片失败,param:%v，err:%v", param, err))
			response.ResError(ctx, "generate interactive image error")
			return
		}
		response.ResSuccess(ctx, res)
	} else if bizType == 3 {
		res, err := c.ImageService.GetTemplateImageId(ctx, param)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GenerateImages]，生成模板图片失败,param:%v，err:%v", param, err))
			response.ResError(ctx, "generate template image error")
			return
		}
		response.ResSuccess(ctx, res)
	} else if bizType == 4 {
		paths := param.FilePaths          // 假设这里是一个包含100个文件路径的切片
		results := make(chan string, 100) // 创建一个channel来收集结果
		errors := make(chan error, 100)   // 创建一个channel来收集错误

		for _, path := range paths {
			go func(p string) {
				imageData, err := ioutil.ReadFile(p)
				if err != nil {
					errors <- err
					return
				}
				res, err := c.ImageService.GenerateImageAndUpload2S3WithBytes(imageData, ctx, param, "")
				if err != nil {
					errors <- err
					return
				}
				results <- res
			}(path)
		}

		// 等待所有goroutine完成
		go func() {
			for i := 0; i < cap(results); i++ {
				if err := <-errors; err != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GenerateImages]，生成互动图片失败,param:%v，err:%v", param, err))
					response.ResError(ctx, "generate interactive image error")
					return
				}
				if res := <-results; res != "" {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GenerateImages]，result:%s", res))
				}
			}
			// 所有结果处理完毕后，返回成功响应
			response.ResSuccess(ctx, "All images uploaded successfully")
		}()
	}
}

func (c ImageController) GetShortUrl(ctx *gin.Context) {
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("GetShortUrl，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	shortUrlParam := &request.ShortUrlParam{}
	ctx.ShouldBindJSON(shortUrlParam)
	scene := shortUrlParam.Scene
	var shortDto dto.ShortDto
	m := orderedmap.New()
	if scene == 1 {

		//mcgg
		urls := shortUrlParam.LongUrls
		for _, url := range urls {
			shortDto = dto.ShortDto{
				LongUrl:         url,
				ActivityId:      "mcgg2025wa",
				ProjectId:       config.ApplicationConfig.Wa.McggShortProject,
				ShortLinkGenUrl: config.ApplicationConfig.Wa.McggShortLinkGenUrl,
				ShortLinkPrefix: config.ApplicationConfig.Wa.McggShortLinkPrefix,
				SignKey:         config.ApplicationConfig.Wa.McggShortLinkSignKey,
			}
			res, err := c.ShortUrlService.GetShortUrl(ctx, shortDto)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GetShortUrl]，生成短链接失败,shortUrlParam:%v，err:%v", shortUrlParam, err))
				response.ResError(ctx, "generate short url error")
				return
			}
			m.Set(url, res)
		}
	} else {
		//mlbb
		urls := shortUrlParam.LongUrls
		for _, url := range urls {
			shortDto = dto.ShortDto{
				LongUrl:         url,
				ActivityId:      "mlbbmy",
				ProjectId:       config.ApplicationConfig.Wa.MlbbShortProject,
				ShortLinkGenUrl: config.ApplicationConfig.Wa.MlbbShortLinkGenUrl,
				ShortLinkPrefix: config.ApplicationConfig.Wa.MlbbShortLinkPrefix,
				SignKey:         config.ApplicationConfig.Wa.MlbbShortLinkSignKey,
			}
			res, err := c.ShortUrlService.GetShortUrl(ctx, shortDto)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[GetShortUrl]，生成短链接失败,shortUrlParam:%v，err:%v", shortUrlParam, err))
				response.ResError(ctx, "generate short url error")
				return
			}
			m.Set(url, res)
		}
	}
	response.ResSuccess(ctx, m)
}

func (c ImageController) RandomMessage(ctx *gin.Context) {
	interactive := &nx.Interactive{}
	ctx.ShouldBindJSON(interactive)

	// 每批次处理12个请求
	batchSize := 12
	totalBatches := 1000
	rand.Seed(time.Now().UnixNano()) // 设置随机种子

	for batch := 0; batch < totalBatches; batch++ {
		// 启动一个批次的goroutine
		var wg sync.WaitGroup
		strings := make([]string, 0)

		// 动态生成昵称列表
		for i := 0; i < batchSize; i++ {
			strings = append(strings, fmt.Sprintf("助力人昵称测试%d-%d", batch, i))
		}

		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			j := j
			go func(index int) {
				defer wg.Done()
				imageParam := &request.SynthesisParam{
					BizType:         2,
					LangNum:         "01",
					NicknameList:    strings,
					CurrentProgress: int64(rand.Intn(8)),
				}
				url, err := c.ImageService.GetInteractiveImageUrl(ctx, imageParam, "")
				if err != nil {
					log.Printf("方法[GenerateImages]，生成互动图片失败, param:%v，err:%v\n", imageParam, err)
					return
				}
				interactive.Header.Image.Link = url
				service.GetRanDomMessage(ctx, interactive, batch, j) // 并发执行
			}(j)
		}
		wg.Wait() // 等待当前批次的所有goroutine完成
		fmt.Printf("batch all done %d 完成\n", batch)
		time.Sleep(1 * time.Second) // 每个批次完成后暂停1秒
	}
}

// 包含数字1, 2, 3, 4, 6, 7, 8的数组
var numbers = []int{1, 2, 3, 4, 6, 7, 8}
