package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util/config/encoder/json"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
	"unicode/utf8"
)

type ImageService struct {
}

const BASE_PATH = "./resources/image/"

var langMap = map[string]string{
	"01": "zh_CN",
	"02": "en",
	"03": "bm",
	"04": "id",
}

// 得到互动消息的图片url
func (u ImageService) GetInteractiveImageUrl(ctx *gin.Context, req *request.SynthesisParam, waId string) (string, error) {
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始调用生成图片方法,waId:%v,req:%v", waId, req))

	if req.NicknameList == nil || len(req.NicknameList) == 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("DealImage方法，result：%v", "nickname is empty"))
		return "", errors.New("nickname is empty")
	}
	if req.CurrentProgress < 0 || req.CurrentProgress > 8 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("DealImage方法，result：%v currentProgress:%v", "currentProgress is error", req.CurrentProgress))
		return "", errors.New("currentProgress is error")
	}
	progress := req.CurrentProgress
	lang := langMap[req.LangNum]
	if progress == 5 {
		return langFixedCover[lang][progress], nil
	}
	imageData, err := u.CreateProgressCoverReturnBytes(ctx, req, waId)
	if err != nil {
		return "", errors.New("create progress cover error")
	}
	compressImage, err := compressImage(imageData, 70)
	if err != nil {
		return "", errors.New("compressImage cover error")
	}

	imageUrl, err := u.GenerateImageAndUpload2S3WithBytes(compressImage, ctx, req, waId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("DealImage方法，result：%v err:%v", "generate image error", err))
		return "", errors.New("generate image error")
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("完成调用生成图片方法,waId:%v,req:%v", waId, req))
	return imageUrl, nil
}

// 得到模板消息的图片id
func (u ImageService) GetTemplateImageId(ctx *gin.Context, req *request.SynthesisParam) (string, error) {

	imageId, err := u.GenerateImageAndUpload2NX(ctx, req)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("DealImage方法，result：%v err:%v", "generate image error", err))
		return "", errors.New("generate image error")
	}
	return imageId, nil
}

func (u ImageService) GenerateImageAndUpload2S3WithBytes(imageData []byte, ctx context.Context, req *request.SynthesisParam, waId string) (string, error) {
	// 生成10个字符的随机字符串
	randomString := generateRandomString(10)

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-获取预签名-开始,waId:%v,req:%v", waId, req))

	// 组合时间戳、随机字符串和文件后缀
	fileName := fmt.Sprintf("%d-%s.png", time.Now().UnixNano()/int64(time.Millisecond), randomString)

	//获取预签名URL

	preSignParam := map[string]string{
		//"bucket": config.ApplicationConfig.S3Config.Bucket,
		"key": fileName,
	}

	preSignUrl, err := getPreSignUrl(preSignParam)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "get preSignUrl error", err))
		return "", errors.New("get preSignUrl error")
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-获取预签名-结束,waId:%v,req:%v", waId, req))
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-上传文件-开始,waId:%v,req:%v", waId, req))

	err = putObject2S3(preSignUrl, imageData)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "upload to s3 error", err))
		return "", errors.New("pre sign url upload to s3 error")
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-上传文件-结束,waId:%v,url:%v,req:%v", waId, config.ApplicationConfig.S3Config.DonAmin+fileName, req))

	return config.ApplicationConfig.S3Config.DonAmin + fileName, nil
}

func (u ImageService) GenerateImageAndUpload2S3(coverFilePath string, ctx context.Context, req *request.SynthesisParam) (string, error) {

	// 读取图片文件
	imageData, err := ioutil.ReadFile(coverFilePath)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "read file error", err))
		return "", errors.New("read file error")
	}

	// 生成10个字符的随机字符串
	randomString := generateRandomString(10)

	// 组合时间戳、随机字符串和文件后缀
	fileName := fmt.Sprintf("%d-%s.png", time.Now().UnixNano()/int64(time.Millisecond), randomString)

	//获取预签名URL

	preSignParam := map[string]string{
		//"bucket": config.ApplicationConfig.S3Config.Bucket,
		"key": fileName,
	}

	preSignUrl, err := getPreSignUrl(preSignParam)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "get preSignUrl error", err))
		return "", errors.New("get preSignUrl error")
	}

	err = putObject2S3(preSignUrl, imageData)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "upload to s3 error", err))
		return "", errors.New("pre sign url upload to s3 error")
	}

	// 删除临时文件
	err = os.Remove(coverFilePath)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "delete file error", err))
		return "", errors.New("delete file error")
	}
	return config.ApplicationConfig.S3Config.DonAmin + fileName, nil
}

var fontBytesMap = make(map[string][]byte)
var boldFontBytesMap = make(map[string][]byte)

var langFont = map[string]string{
	"zh_CN": "./resources/image/zh_CN/AlibabaPuHuiTi-3-55-Regular.ttf",
	"en":    "./resources/image/en/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"bm":    "./resources/image/id/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"id":    "./resources/image/id/AlibabaPuHuiTi-3-75-SemiBold.ttf",
}

func getFont(lang string) ([]byte, error) {
	if fontBytesMap[lang] != nil {
		return fontBytesMap[lang], nil
	}
	fontBytes, err := os.ReadFile(langFont[lang])
	if err != nil {
		return nil, err
	}
	fontBytesMap[lang] = fontBytes
	return fontBytes, nil
}

var httpTransportClient = http.Client{}

var preSignClient = http.Client{}

func InitImageService() {
	for lang, fontPath := range langFont {
		fontBytes, err := os.ReadFile(fontPath)
		if err != nil {
			// 处理错误，例如记录日志或者panic
			panic(err) // 这里选择panic，因为初始化失败是严重错误
		}
		fontBytesMap[lang] = fontBytes
	}

	for lang, fontPath := range langBoldFont {
		fontBytes, err := os.ReadFile(fontPath)
		if err != nil {
			// 处理错误，例如记录日志或者panic
			panic(err) // 这里选择panic，因为初始化失败是严重错误
		}
		boldFontBytesMap[lang] = fontBytes
	}

	httpTransportClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,              // 最大空闲连接数
			IdleConnTimeout:    10 * time.Second, // 空闲连接超时时间
			MaxConnsPerHost:    200,
			DisableCompression: true, // 禁用压缩，因为压缩和解压缩会消耗CPU资源
		},
	}

	preSignClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,              // 最大空闲连接数
			IdleConnTimeout:    20 * time.Second, // 空闲连接超时时间
			MaxConnsPerHost:    200,
			DisableCompression: true, // 禁用压缩，因为压缩和解压缩会消耗CPU资源
		},
	}
}

var langBoldFont = map[string]string{
	"zh_CN": "./resources/image/zh_CN/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"en":    "./resources/image/en/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"bm":    "./resources/image/id/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"id":    "./resources/image/id/AlibabaPuHuiTi-3-75-SemiBold.ttf",
}

func getBoldFont(lang string) ([]byte, error) {
	if fontBytesMap[lang] != nil {
		return fontBytesMap[lang], nil
	}
	fontBytes, err := os.ReadFile(langBoldFont[lang])
	if err != nil {
		return nil, err
	}
	fontBytesMap[lang] = fontBytes
	return fontBytes, nil
}

var coverCopywriting = map[string]map[string]string{
	"zh_CN": {
		"left":  "你的好友【",
		"right": "】为你助力成功！",
	},
	"en": {
		"left":  "YOUR FRIEND, [",
		"right": "] HAS SUCCESSFULLY ASSISTED YOU!",
	},
	"id": {
		"left":  "Temanmu [",
		"right": "], BERHASIL MEMBANTU KAMU!",
	},
}

func truncateString(s string, maxRunes int) string {
	// 将字符串转换为rune切片，以正确处理UTF-8字符
	runes := []rune(s)

	// 如果rune切片的长度小于或等于maxRunes，则直接返回原始字符串
	if len(runes) <= maxRunes {
		return s
	}

	// 截取前maxRunes个rune，并转换回字符串
	truncated := string(runes[:maxRunes])

	// 如果截取后的字符串长度小于原始字符串长度，追加"..."
	if utf8.RuneCountInString(truncated) < utf8.RuneCountInString(s) {
		truncated += "..."
	}

	return truncated
}

var langFixedCover = map[string]map[int64]string{
	"zh_CN": {
		5: "https://akmweb.outweb.mc-gogo.com/1737104620349-mHuyE0zf1k.png",
	},
	"en": {
		5: "https://akmweb.outweb.mc-gogo.com/1737104620349-mHuyE0zf1k.png",
	},
	"id": {
		5: "https://akmweb.outweb.mc-gogo.com/1737014648949-EPXMQ9kpM3.png",
	},
}

func (u ImageService) CreateProgressCoverReturnBytes(ctx context.Context, req *request.SynthesisParam, waId string) ([]byte, error) {
	lang := langMap[req.LangNum]
	if lang == "id" || lang == "en" {
		return u.CreateProgressCoverWithOTFReturnBytes(ctx, req)
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-读取本地图片-开始,waId:%v,req:%v", waId, req))

	file, err := os.Open(BASE_PATH + lang + "/banner" + fmt.Sprintf("%d", req.CurrentProgress) + ".jpg")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "open file error", err))
		return nil, err
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "decode jpeg error", err))
		return nil, err
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-读取本地图片-结束,waId:%v,req:%v", waId, req))

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-合成图片-开始,waId:%v,req:%v", waId, req))

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	nicknames := req.NicknameList
	nickname := nicknames[req.CurrentProgress-1]
	nickname = truncateString(nickname, 8)

	leftText := coverCopywriting[lang]["left"]
	rightText := coverCopywriting[lang]["right"]

	regularFontBytes, err := getFont(lang)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "font file error", err))
		return nil, err
	}
	boldFontBytes, err := getBoldFont(lang)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "bold font file error", err))
		return nil, err
	}

	regularFont, err := freetype.ParseFont(regularFontBytes)
	if err != nil {
		return nil, err
	}
	boldFont, err := freetype.ParseFont(boldFontBytes)
	if err != nil {
		return nil, err
	}

	fontSize := 36.0
	fixedColor := image.NewUniform(color.RGBA{R: 131, G: 37, B: 3, A: 255})
	nicknameColor := image.NewUniform(color.RGBA{R: 255, G: 30, B: 30, A: 255})

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	centerX, centerY := 570, 355

	leftFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	rightFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})

	drawer := font.Drawer{
		Face: leftFace,
	}
	leftWidth := drawer.MeasureString(leftText).Ceil()

	drawer.Face = boldFace
	nicknameWidth := drawer.MeasureString(nickname).Ceil()

	drawer.Face = rightFace
	rightWidth := drawer.MeasureString(rightText).Ceil()

	totalWidth := leftWidth + nicknameWidth + rightWidth

	startX := centerX - totalWidth/2
	startY := centerY + int(c.PointToFixed(fontSize)>>6)/2

	c.SetFont(regularFont)
	c.SetFontSize(fontSize)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(leftText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += leftWidth
	c.SetFont(boldFont)
	c.SetSrc(nicknameColor)
	_, err = c.DrawString(nickname, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += nicknameWidth
	c.SetFont(regularFont)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(rightText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		return nil, err
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("调用生成图片方法-合成图片-结束,waId:%v,req:%v", waId, req))

	return buf.Bytes(), nil
}

func (u ImageService) CreateProgressCoverWithOTFReturnBytes(ctx context.Context, req *request.SynthesisParam) ([]byte, error) {
	lang := langMap[req.LangNum]
	file, err := os.Open(BASE_PATH + lang + "/banner" + fmt.Sprintf("%d", req.CurrentProgress) + ".jpg")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "open file error", err))
		return nil, err
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "decode jpeg error", err))
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	nicknames := req.NicknameList
	nickname := nicknames[req.CurrentProgress-1]
	nickname = truncateString(nickname, 8)

	leftText := coverCopywriting[lang]["left"]
	rightText := coverCopywriting[lang]["right"]

	regularFontBytes, err := getFont(lang)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "font file error", err))
		return nil, err
	}
	boldFontBytes, err := getBoldFont(lang)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "bold font file error", err))
		return nil, err
	}

	regularFont, err := freetype.ParseFont(regularFontBytes)
	if err != nil {
		return nil, err
	}
	boldFont, err := freetype.ParseFont(boldFontBytes)
	if err != nil {
		return nil, err
	}

	fontSize := 30.0
	fixedColor := image.NewUniform(color.RGBA{R: 131, G: 37, B: 3, A: 255})
	nicknameColor := image.NewUniform(color.RGBA{R: 255, G: 30, B: 30, A: 255})

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	centerX, centerY := 570, 355

	leftFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	rightFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})

	drawer := font.Drawer{
		Face: leftFace,
	}
	leftWidth := drawer.MeasureString(leftText).Ceil()

	drawer.Face = boldFace
	nicknameWidth := drawer.MeasureString(nickname).Ceil()

	drawer.Face = rightFace
	rightWidth := drawer.MeasureString(rightText).Ceil()

	totalWidth := leftWidth + nicknameWidth + rightWidth

	startX := centerX - totalWidth/2
	startY := centerY + int(c.PointToFixed(fontSize)>>6)/2

	c.SetFont(regularFont)
	c.SetFontSize(fontSize)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(leftText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += leftWidth
	c.SetFont(boldFont)
	c.SetSrc(nicknameColor)
	_, err = c.DrawString(nickname, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += nicknameWidth
	c.SetFont(regularFont)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(rightText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 压缩图片并返回压缩后的图片数据
func compressImage(imageData []byte, quality int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	// 创建一个新的bytes.Buffer用于存储压缩后的图片数据
	var buf bytes.Buffer
	// 使用JPEG格式压缩图片，并设置压缩质量
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (u ImageService) CreateProgressCoverWithOTF(ctx context.Context, req *request.SynthesisParam) (string, error) {
	// Open the base image file
	lang := langMap[req.LangNum]
	file, err := os.Open(BASE_PATH + lang + "/banner" + fmt.Sprintf("%d", req.CurrentProgress) + ".jpg")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "open file error", err))
		return "", errors.New("open file error")
	}
	defer file.Close()
	var img image.Image
	img, err = jpeg.Decode(file)

	// Create a new RGBA image for drawing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	// Get the nickname
	nicknames := req.NicknameList
	nickname := nicknames[req.CurrentProgress-1]
	nickname = truncateString(nickname, 8)

	// Left and right fixed copywriting
	leftText := coverCopywriting[lang]["left"]   // e.g., "YOUR FRIEND ["
	rightText := coverCopywriting[lang]["right"] // e.g., "] HAVE SUCCESSFULLY SUPPORTED YOU!"

	// Load fonts
	regularFontBytes, err := getFont(lang) // Font for left and right fixed copywriting
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "font file error", err))
		return "", errors.New("font file error")
	}
	boldFontBytes, err := getBoldFont(lang) // Bold font for the nickname
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2S3，result：%v err:%v", "bold font file error", err))
		return "", errors.New("bold font file error")
	}

	regularFont, err := freetype.ParseFont(regularFontBytes)
	if err != nil {
		return "", errors.New("parse font error for regular font")
	}
	boldFont, err := freetype.ParseFont(boldFontBytes)
	if err != nil {
		return "", errors.New("parse font error for bold font")
	}

	// Define font size and colors
	fontSize := 23.0
	fixedColor := image.NewUniform(color.RGBA{R: 131, G: 37, B: 3, A: 255})
	nicknameColor := image.NewUniform(color.RGBA{R: 255, G: 30, B: 30, A: 255})

	// Create a freetype context
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	// Define text positions
	centerX, centerY := 570, 355

	// Measure text width
	leftFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	rightFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})

	drawer := font.Drawer{
		Face: leftFace,
	}
	leftWidth := drawer.MeasureString(leftText).Ceil()

	drawer.Face = boldFace
	nicknameWidth := drawer.MeasureString(nickname).Ceil()

	drawer.Face = rightFace
	rightWidth := drawer.MeasureString(rightText).Ceil()

	totalWidth := leftWidth + nicknameWidth + rightWidth

	// Calculate the starting coordinates
	startX := centerX - totalWidth/2
	startY := centerY + int(c.PointToFixed(fontSize)>>6)/2

	// Draw left copywriting
	c.SetFont(regularFont)
	c.SetFontSize(fontSize)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(leftText, freetype.Pt(startX, startY))
	if err != nil {
		return "", errors.New("draw left text error")
	}

	// Update startX for nickname
	startX += leftWidth

	// Draw nickname
	c.SetFont(boldFont)
	c.SetFontSize(fontSize)
	c.SetSrc(nicknameColor)
	_, err = c.DrawString(nickname, freetype.Pt(startX, startY))
	if err != nil {
		return "", errors.New("draw nickname error")
	}

	// Update startX for right text
	startX += nicknameWidth

	// Draw right copywriting
	c.SetFont(regularFont)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(rightText, freetype.Pt(startX, startY))
	if err != nil {
		return "", errors.New("draw right text error")
	}

	// Save the output image
	tmpPath := BASE_PATH + generateRandomString(10) + ".png"
	outFile, err := os.Create(tmpPath)
	if err != nil {
		return "", errors.New("error creating output file")
	}
	defer outFile.Close()

	err = png.Encode(outFile, rgba)
	if err != nil {
		return "", errors.New("error encoding image")
	}
	return tmpPath, nil
}

func (u ImageService) GenerateImageAndUpload2NX(ctx *gin.Context, req *request.SynthesisParam) (string, error) {
	coverFilePath := req.FilePath
	waMsgService := GetWaMsgService()
	fileId, err := waMsgService.UploadMedia2NX(ctx, coverFilePath)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2NX，result：%v err:%v", "upload image to NX error", err))
		return "", errors.New("upload image to NX error")
	}
	// 删除临时文件
	err = os.Remove(coverFilePath)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("generateImageAndUpload2NX，result：%v err:%v", "delete file error", err))
		return "", errors.New("delete file error")
	}
	return fileId, nil
}

const RandomStr = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 生成随机的字母和数字序列
func generateRandomString(n int) string {
	var letters = []rune(RandomStr)
	s := make([]rune, n)
	for i := range s {
		b, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		s[i] = letters[b.Int64()]
	}
	return string(s)
}

func (u ImageService) GeneratePreSignedURL(ctx *gin.Context, request *request.PreSignParam) (string, error) {
	// 创建 AWS 会话
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.ApplicationConfig.S3Config.Region),
		Credentials: credentials.NewStaticCredentials(config.ApplicationConfig.S3Config.AccessKeyID, config.ApplicationConfig.S3Config.SecretAccessKey, ""),
	})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("GeneratePreSignedURL，result：%v err:%v", "create AWS session error", err))
		return "", err
	}

	// 创建 S3 服务客户端
	svc := s3.New(sess)

	// 生成预签名 URL
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(config.ApplicationConfig.S3Config.Bucket),
		Key:    aws.String(request.Key),
	})

	// 设置预签名 URL 的有效期
	urlStr, err := req.Presign(15 * time.Minute) // URL 有效期为 15 分钟
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("GeneratePreSignedURL，result：%v err:%v", "generate presigned URL error", err))
		return "", err
	}
	return urlStr, nil
}

func getPreSignUrl(bodyData map[string]string) (string, error) {
	bodyBytes, err := json.NewEncoder().Encode(bodyData)
	if err != nil {
		logTracing.LogPrintfP("Error encoding body data: %v", err)
		return "", errors.New(fmt.Sprintf("Error encoding body data: %v", err))
	}
	// 创建请求
	req, err := http.NewRequest("POST", config.ApplicationConfig.S3Config.PreSignUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logTracing.LogPrintfP("Error creating request: %v", err)
		return "", errors.New(fmt.Sprintf("Error creating request: %v", err))
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := preSignClient
	resp, err := client.Do(req)
	if err != nil {
		logTracing.LogPrintfP("Error sending request: %v", err)
		return "", errors.New(fmt.Sprintf("Error sending request: %v", err))
	}
	defer resp.Body.Close()
	// 读取响应体
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resultResponse := &response.ResultResponse{}
	err = json.NewEncoder().Decode(responseBody, resultResponse)
	if err != nil {
		logTracing.LogPrintfP("Error decoding response: %v", err)
		return "", errors.New(fmt.Sprintf("Error decoding response: %v", err))
	}
	if resultResponse.Code != 200 {
		logTracing.LogWarn(context.Background(), logTracing.WarnLogFmt, fmt.Sprintf("调用getPreSignUrl,返回结果失败，报错信息: %v", resultResponse.Message))
		return "", errors.New(fmt.Sprintf("调用getPreSignUrl,返回结果失败，报错信息: %v", resultResponse.Message))
	}
	return resultResponse.Data.(string), nil
}

// PutObject2S3 使用预签名 URL 上传文件到 S3
func putObject2S3(preSignUrl string, fileData []byte) error {
	// 使用预签名 URL 上传文件
	req, err := http.NewRequest("PUT", preSignUrl, bytes.NewReader(fileData))
	if err != nil {
		logTracing.LogPrintfP("Failed to create request: %v", err)
		return errors.New(fmt.Sprintf("Failed to create request: %v", err))
	}

	req.Header.Set("Content-Type", "image/png") // 设置文件类型

	client := httpTransportClient
	resp, err := client.Do(req)
	if err != nil {
		logTracing.LogPrintfP("Failed to upload file: %v", err)
		return errors.New(fmt.Sprintf("Failed to upload file: %v", err))
	}
	defer resp.Body.Close()

	// 检查上传响应
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Failed to upload file, status: %s, body: %s", resp.Status, body)
		return errors.New(fmt.Sprintf("Failed to upload file, status: %s, body: %s", resp.Status, body))
	}
	return nil
}
