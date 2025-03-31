package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util/config/encoder/json"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

func IsOverRedisCdkLen(ctx context.Context) (bool, error) {
	// redis
	template := redis_template.NewRedisTemplate()
	cdkTypeList := constant.GetAllCdkType()

	for _, cdkType := range cdkTypeList {

		cdkKey := constant.GetCdkKey(config.ApplicationConfig.Activity.Id, cdkType)

		code, err := template.Exists(ctx, cdkKey)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("IsOverRedisCdkLen方法，查询%v list长度报错,err：%v", cdkType, err))
			return false, errors.New(fmt.Sprintf("IsOverRedisCdkLen方法，查询%v list长度报错,err：%v", cdkType, err))
		}
		if code == 0 {
			return true, nil
		}
	}
	return false, nil
}

func IsUnderPercentCdkLen(ctx context.Context, underPercent float64) (bool, error) {
	methodName := "IsUnderPercentCdkLen"
	// redis
	template := redis_template.NewRedisTemplate()
	cdkTypeList := constant.GetAllCdkType()

	for _, cdkType := range cdkTypeList {

		cdkKey := constant.GetCdkKey(config.ApplicationConfig.Activity.Id, cdkType)

		code, err := template.Exists(ctx, cdkKey)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v list长度报错,err：%v", methodName, cdkType, err))
			return false, errors.New(fmt.Sprintf("方法[%s]，查询%v list长度报错,err：%v", methodName, cdkType, err))
		}
		if code == 0 {
			return true, nil
		}

		cdkNotUsedLen, err := template.LLen(ctx, cdkKey).Result()
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取%v存量的长度失败，key:%v", methodName, cdkType, cdkKey))
			return false, err
		}

		cdkInfoKey := constant.GetCdkInfoKey(config.ApplicationConfig.Activity.Id, cdkType)
		cdkInfoStr, err := template.Get(ctx, cdkInfoKey)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取%v总长度失败，key:%v", methodName, cdkType, cdkInfoKey))
			return false, err
		}

		cdkInfo := &response.CdkInfo{}
		err = json.NewEncoder().Decode([]byte(cdkInfoStr), cdkInfo)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，cdkInfo转实体报错,cdkInfoStr:%v,err：%v", methodName, cdkInfoStr, err))
			return false, err
		}

		cdkCount := cdkInfo.CdkCount

		percent := math.Round(float64(cdkCount-cdkNotUsedLen)/float64(cdkCount)*10000) / 100
		if percent > underPercent {
			return true, nil
		}
	}
	return false, nil
}

func GetCdkByCdkType(ctx context.Context, cdkType string) (string, bool, error) {
	methodName := "GetCdkByCdkType"
	// redis
	template := redis_template.NewRedisTemplate()

	cdkKey := constant.GetCdkKey(config.ApplicationConfig.Activity.Id, cdkType)

	code, err := template.Exists(ctx, cdkKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v list长度报错,err：%v", methodName, cdkType, err))
		return "", false, err
	}
	if code == 0 {
		return "", false, nil
	}

	cdk, err := template.BRPop(ctx, cdkKey, constant.CdkKeyTimeOut)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取cdk失败,cdkType:%v,err:%v", methodName, cdkType, err))
		return "", true, err
	}

	return cdk[1], true, nil
}

func InitHelpWeight(ctx context.Context) {
	methodName := "InitHelpWeight"

	template := redis_template.NewRedisTemplate()
	lockKey := constant.GetHelpTextLockKey(config.ApplicationConfig.Activity.Id)

	getLock, err := template.SetNX(context.Background(), lockKey, "1", time.Second*60).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx失败，本实例不执行任务", methodName))
		return
	}
	if !getLock {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取分布式锁失败，本实例不执行任务", methodName))
		return
	}

	defer func() {
		del := template.Del(context.Background(), lockKey)
		if !del {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败", methodName))
		}
	}()

	helpTextWeightKey := constant.GetHelpTextWeightKey(config.ApplicationConfig.Activity.Id)
	helpTextList := config.ApplicationConfig.Activity.HelpTextList
	paramsBytes, err := json.NewEncoder().Encode(helpTextList)
	if err != nil {
		//logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，HelpTextList转换json失败,HelpTextList:%v,err:%v", methodName, helpTextList, err))
		log.Fatal(fmt.Sprintf("方法[%s]，HelpTextList转换json失败,HelpTextList:%v,err:%v", methodName, helpTextList, err))
		return
	}
	paramsStr := string(paramsBytes)
	set := template.Set(ctx, helpTextWeightKey, paramsStr)
	if !set {
		log.Fatal(fmt.Sprintf("方法[%s]，查询%v 初始化help权重报错,err：%v", methodName, helpTextWeightKey, err))
		return
	}
	return
}

// GetHelpTextWeight 获取权重的
func GetHelpTextWeight(ctx *gin.Context) (*config.HelpText, error) {
	ctx2 := context.Background()
	methodName := "GetHelpTextWeight"
	template := redis_template.NewRedisTemplate()

	helpTextWeightKey := constant.GetHelpTextWeightKey(config.ApplicationConfig.Activity.Id)

	code, err := template.Exists(ctx2, helpTextWeightKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v 是否存在报错,err：%v", methodName, helpTextWeightKey, err))
		return nil, errors.New(fmt.Sprintf("方法[%s]，查询%v 是否存在报错,err：%v", methodName, helpTextWeightKey, err))
	}
	if code == 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v helpText数据不存在,err：%v", methodName, helpTextWeightKey, err))
		return nil, errors.New(fmt.Sprintf("方法[%s]，查询%v helpText数据不存在,err：%v", methodName, helpTextWeightKey, err))
	}

	str, err := template.Get(ctx2, helpTextWeightKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v 报错,err：%v", methodName, helpTextWeightKey, err))
		return nil, errors.New(fmt.Sprintf("方法[%s]，查询%v 报错,err：%v", methodName, helpTextWeightKey, err))
	}

	helpTextList := make([]*config.HelpText, 0)
	err = json.NewEncoder().Decode([]byte(str), &helpTextList)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，HelpTextList转换json失败,HelpTextList:%v,err:%v", methodName, helpTextList, err))
		return nil, errors.New(fmt.Sprintf("方法[%s]，HelpTextList转换json失败,HelpTextList:%v,err:%v", methodName, helpTextList, err))
	}
	if len(helpTextList) <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，HelpTextList长度为0,HelpTextList:%v,err:%v", methodName, helpTextList, err))
		return nil, errors.New(fmt.Sprintf("方法[%s]，HelpTextList长度为0,HelpTextList:%v,err:%v", methodName, helpTextList, err))
	}

	// 计算权重的累积和
	var totalWeight int
	for _, helpText := range helpTextList {
		totalWeight += helpText.Weight
	}

	// 生成一个随机数，范围从 0 到 totalWeight
	rand.Seed(time.Now().UnixNano()) // 随机种子
	randomWeight := rand.Intn(totalWeight)

	// 根据随机数选择对应的值
	for _, helpText := range helpTextList {
		randomWeight -= helpText.Weight
		if randomWeight <= 0 {
			return helpText, nil
		}
	}

	return nil, errors.New("没有任何值")
}

func SAddKey(methodName string, key string, value string) (int64, error) {
	ctx := context.Background()
	// 给非白拦截redis增加phone
	template := redis_template.NewRedisTemplate()
	newCount, err := template.SAdd(context.Background(), key, value).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，增加%v 报错,err：%v", methodName, key, err))
		return 0, errors.New(fmt.Sprintf("方法[%s]，增加%v，报错,err：%v", methodName, key, err))
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，增加%v 完成, 增加数量addCount：%v", methodName, key, newCount))
	return newCount, nil
}

func AddIncrKey(methodName string, key string) (int64, error) {
	ctx := context.Background()
	// 给非白拦截redis增加次数
	template := redis_template.NewRedisTemplate()
	newCount, err := template.Incr(context.Background(), key).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，增加%v 报错,err：%v", methodName, key, err))
		return 0, errors.New(fmt.Sprintf("方法[%s]，增加%v，报错,err：%v", methodName, key, err))
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，增加%v 成功,newCount：%v", methodName, key, newCount))
	return newCount, nil
}

// GetAllDaysIncrKeyCount 获取统计数
func GetAllDaysIncrKeyCount(methodName string, format, today string, template redis_template.RedisTemplate) (int64, error) {
	//ctx := context.Background()

	//mapper := dao.GetReportMsgInfoMapper()
	//dateList, err := mapper.SelectDays(config.ApplicationConfig.Activity.Id)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询之前的统计日期报错,err：%v", methodName, err))
	//	return 0, err
	//}
	//dateList = append(dateList, today)

	dateList := get30Days(today)
	dateList = append(dateList, today)
	allCount := int64(0)
	for _, date := range dateList {
		for _, channel := range config.ApplicationConfig.Activity.ChannelList {
			for _, language := range config.ApplicationConfig.Activity.LanguageList {
				date = constant.ReplaceChineseMonthDay(date)
				key := fmt.Sprintf(format, config.ApplicationConfig.Activity.Id, date, channel, language)
				count, err := GetIncrKeyCount(methodName, key, template)
				if err != nil {
					return 0, err
				}
				allCount += count
			}
		}
	}

	return allCount, nil
}

// GetIncrKeyCount 获取统计数
func GetIncrKeyCount(methodName string, key string, template redis_template.RedisTemplate) (int64, error) {
	ctx := context.Background()
	//template := redis_template.NewRedisTemplate()

	exists, err := template.Exists(ctx, key)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v 是否存在报错,err：%v", methodName, key, err))
		return 0, err
	}
	if exists == 0 {
		return 0, nil
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("获取redis统计开始，key：%v", key))
	countStr, err := template.Get(context.Background(), key)
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("获取redis统计结束，key：%v", key))
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v 报错,err：%v", methodName, key, err))
		return 0, err
	}
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%v]，字符串转int失败:%v", methodName, err))
		return 0, err
	}
	return count, nil
}

// GetAllIncrKeyCount 获取统计数
func GetAllIncrKeyCount(methodName string, keyFormat string, template redis_template.RedisTemplate) (int64, error) {

	all := int64(0)
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			key := fmt.Sprintf(keyFormat, strconv.Itoa(config.ApplicationConfig.Activity.Id), channel, language)
			count, err := GetIncrKeyCount(methodName, key, template)
			if err != nil {
				return 0, err
			}
			all += count
		}
	}
	return all, nil
}

func AddHelpInfoCache(methodName string, key string, values *dto.HelpCacheDto) (int, []*dto.HelpCacheDto, error) {
	ctx := context.Background()

	valuesBytes, err := json.NewEncoder().Encode(values)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换HelpCacheDto为json失败,err:%v", methodName, err))
		return 0, nil, err
	}

	// redis增加次数
	template := redis_template.NewRedisTemplate()
	newCount, err := template.RPush(context.Background(), key, string(valuesBytes)).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，增加%v 报错,err：%v", methodName, key, err))
		return 0, nil, err
	}
	if newCount > int64(config.ApplicationConfig.Activity.Stage3Award.HelpNum) {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，用户已助力满不需要再次助力。", methodName))
		// 已超出
		RemoveHelpInfoCache(methodName, key)
		return int(newCount), nil, nil
	}
	helpCacheDtoList, err := QueryHelpInfoCache(methodName, key, newCount)
	if err != nil {
		return 0, nil, err
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，增加%v 成功,newCount：%v", methodName, key, newCount))
	return int(newCount), helpCacheDtoList, nil
}

func RemoveHelpInfoCache(methodName string, key string) error {
	ctx := context.Background()
	// redis减少次数
	template := redis_template.NewRedisTemplate()
	_, err := template.RPop(context.Background(), key).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，减少%v 报错,err：%v", methodName, key, err))
		return err
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，减少%v 成功", methodName, key))
	return nil
}

func QueryHelpInfoCache(methodName string, key string, newCount int64) ([]*dto.HelpCacheDto, error) {
	ctx := context.Background()
	// 给非白拦截redis增加次数
	template := redis_template.NewRedisTemplate()
	helpCacheStrArray, err := template.LRange(context.Background(), key, 0, newCount-1).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v，start：%v,end:%v 报错,err：%v", methodName, key, 0, newCount-1, err))
		return nil, errors.New(fmt.Sprintf("方法[%s]，查询%v，start：%v,end:%v 报错,err：%v", methodName, key, 0, newCount-1, err))
	}

	helpCacheList := make([]*dto.HelpCacheDto, 0)
	for _, helpCacheStr := range helpCacheStrArray {
		helpCacheDto := &dto.HelpCacheDto{}
		err = json.NewEncoder().Decode([]byte(helpCacheStr), &helpCacheDto)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，helpCacheStrt转换实体失败,helpCacheStr:%v,err:%v", methodName, helpCacheStr, err))
			return nil, errors.New(fmt.Sprintf("方法[%s]，helpCacheStrt转换实体失败,helpCacheStr:%v,err:%v", methodName, helpCacheStr, err))
		}
		helpCacheList = append(helpCacheList, helpCacheDto)
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，查询%v 成功,helpCacheList：%v", methodName, key, helpCacheList))
	return helpCacheList, nil
}

func get30Days(dateStr string) []string {
	var dateStrList []string
	date, err := parseDate(dateStr)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return dateStrList
	}

	dates := make([]time.Time, 30)
	for i := 0; i < 30; i++ {
		dates[i] = date.AddDate(0, 0, -1-i)
	}

	for _, d := range dates {
		dateStrList = append(dateStrList, d.Format("1月2日"))
	}
	return dateStrList
}

func parseDate(dateStr string) (time.Time, error) {
	layout := "1月2日"
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
