package http_client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// 配置SSL
func configureTLS() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, // 如果你需要忽略证书验证
	}
}

var httpClientPool = http.Client{}

func InitHttpClientPool() {
	httpClientPool = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       200,              // 最大空闲连接数
			IdleConnTimeout:    10 * time.Second, // 空闲连接超时时间
			MaxConnsPerHost:    500,
			DisableCompression: true, // 禁用压缩，因为压缩和解压缩会消耗CPU资源
		},
	}

}

// DoPostSSL 执行HTTPS POST请求
func DoPostSSL(apiUrl string, bodyMap any, headers map[string]string, socketTimeout time.Duration, connectTimeout time.Duration) (string, error) {

	bodyBytes, err := json.NewEncoder().Encode(bodyMap)
	if err != nil {
		return "", errors.New(fmt.Sprintf("DoPostSSL,转换json报错，message:%v", bodyMap))
	}
	// 创建请求体
	body := bytes.NewBuffer(bodyBytes)

	// 创建请求
	req, err := http.NewRequest("POST", apiUrl, body)
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "UTF-8")

	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}
	// 创建HTTP客户端
	client := httpClientPool

	// 记录开始时间
	start := util.GetNowCustomTime()
	logTracing.LogPrintfP("请求接口开始：api：%v;startTime: %v", apiUrl, start)
	// 发送POST请求
	resp, err := client.Do(req)

	logTracing.LogPrintfP("请求接口结束：api：%v;endTime: %v;耗时：%v", apiUrl, util.GetNowCustomTime().Time, time.Since(start.Time))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {

		errMsg := fmt.Sprintf("post statusCode error %d, %s, %s", resp.StatusCode, apiUrl, string(bodyBytes))
		return "", fmt.Errorf(errMsg)
	}

	// 读取响应体
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}

// PostWithForm 发送一个包含表单数据的POST请求
func PostWithForm(url string, urlParameters url.Values) (string, error) {
	// 将表单数据编码为字符串
	dataString := urlParameters.Encode()

	// 将数据转换为字节切片
	payload := strings.NewReader(dataString)

	// 创建一个请求
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// 设置请求头信息
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(dataString)))

	// 发送POST请求
	client := &http.Client{}

	// 记录开始时间
	start := util.GetNowCustomTime()
	logTracing.LogPrintfP("请求接口开始：api：%v;startTime: %v", url, start)
	resp, err := client.Do(req)
	logTracing.LogPrintfP("请求接口结束：api：%v;endTime: %v;耗时：%v", url, util.GetNowCustomTime(), time.Since(start.Time))

	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	// 打印请求状态
	fmt.Printf("Response Status: %s\n", resp.Status)
	// 可选：打印响应体
	responseBody, err := ioutil.ReadAll(resp.Body)
	return string(responseBody), err
}
