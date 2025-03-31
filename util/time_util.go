package util

import (
	"errors"
	"fmt"
	"go-fission-activity/config"
	"go-fission-activity/config/initConfig"
	"log"
	"strconv"
	"strings"
	"time"
)

type CustomTime struct {
	time.Time
	IsNotZero bool
}

func (c *CustomTime) UnmarshalJSON(b []byte) error {
	dateStr := string(b) // something like `"2017-08-20"`

	dateStr = strings.ReplaceAll(dateStr, "\"", "")

	if dateStr == "null" || dateStr == "" || dateStr == "\"\"" {
		c.IsNotZero = false
		return nil
	}

	if strings.Contains(dateStr, "T") {

		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return fmt.Errorf("cant parse date: %#v", err)
		}
		c.Time = t
	} else {
		t, err := time.Parse(`2006-01-02 15:04:05`, dateStr)
		if err != nil {
			return fmt.Errorf("cant parse date: %#v", err)
		}
		c.Time = t
	}
	c.IsNotZero = !c.Time.IsZero()
	return nil
}

func (c CustomTime) MarshalJSON() ([]byte, error) {
	json, err := c.Time.MarshalJSON()
	return json, err
}

func (c CustomTime) Unix() int64 {
	// 格式化时间
	format := c.Format("2006-01-02 15:04:05")

	location := GetLocation()
	nowZone, err := time.ParseInLocation("2006-01-02 15:04:05", format, location)
	if err != nil {
		fmt.Errorf("cant parse date: %#v", err)
	}
	return nowZone.Unix()
}

func (c CustomTime) UnixMilli() int64 {
	// 格式化时间
	format := c.Format("2006-01-02 15:04:05")

	location := GetLocation()
	nowZone, err := time.ParseInLocation("2006-01-02 15:04:05", format, location)
	if err != nil {
		fmt.Errorf("cant parse date: %#v", err)
	}
	return nowZone.UnixMilli()
}

var shanghaiLocation *time.Location

func GetLocation() *time.Location {
	var err error
	if shanghaiLocation != nil {
		return shanghaiLocation
	} else {
		shanghaiLocation, err = time.LoadLocation("Asia/Shanghai")
		if err != nil {
			fmt.Errorf("获取时区报错: %#v", err)
		}
	}
	log.Printf("设置时区成功，shanghaiLocation：%v", shanghaiLocation)
	return shanghaiLocation
}

func NewCustomTime(t time.Time) CustomTime {
	// 获取当前地区的时间
	localTime := t.In(GetLocation()) // 使用本地时区
	format := localTime.Format("2006-01-02 15:04:05")
	nowZone, err := time.Parse("2006-01-02 15:04:05", format)
	if err != nil {
		fmt.Errorf("cant parse date: %#v", err)
	}
	return CustomTime{
		Time:      nowZone,
		IsNotZero: !nowZone.IsZero(),
	}
}

func GetNowCustomTime() CustomTime {
	// 获取当前时间（默认是 UTC 时间）
	now := time.Now()
	// 获取当前地区的时间
	localTime := now.In(GetLocation()) // 使用本地时区
	format := localTime.Format("2006-01-02 15:04:05")
	nowZone, err := time.Parse("2006-01-02 15:04:05", format)
	if err != nil {
		fmt.Errorf("cant parse date: %#v", err)
	}
	return CustomTime{
		Time:      nowZone,
		IsNotZero: !localTime.IsZero(),
	}
}

func GetCustomTimeByTime(timestampStr string) (CustomTime, error) {
	// 将字符串转换为 int64
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		fmt.Println("GetCustomTimeByTime,Error parsing timestamp:", err)
		return CustomTime{}, errors.New(fmt.Sprintf("Error parsing timestamp:%v", err))
	}

	// 创建 time.Time 对象
	newTime := time.Unix(timestamp, 0)

	// 获取当前地区的时间
	localTime := newTime.In(time.Local) // 使用本地时区
	format := localTime.Format("2006-01-02 15:04:05")
	nowZone, err := time.Parse("2006-01-02 15:04:05", format)
	if err != nil {
		fmt.Errorf("cant parse date: %#v", err)
	}
	return CustomTime{
		Time:      nowZone,
		IsNotZero: !localTime.IsZero(),
	}, nil
}

func GetTimeOfAfterDays(afterDay int, startTime CustomTime) CustomTime {
	// 获取当前时间后两天的时间
	newTime := startTime.Add(time.Duration(afterDay*24) * time.Hour)
	// 获取当前地区的时间
	//localTime := newTime.In(time.Local) // 使用本地时区
	//format := localTime.Format("2006-01-02 15:04:05")
	//nowZone, err := time.Parse("2006-01-02 15:04:05", format)
	//if err != nil {
	//	fmt.Errorf("cant parse date: %#v", err)
	//}

	return CustomTime{
		Time:      newTime,
		IsNotZero: !newTime.IsZero(),
	}
}

func GetSendRenewMsgTime(afterDay int, startTime CustomTime) CustomTime {
	// 获取当前时间后22小时的时间
	nowZone := startTime.Add(time.Duration(initConfig.GetReFreeFirstHour()) * time.Hour)
	// 获取当前地区的时间
	/*localTime := newTime.In(time.Local) // 使用本地时区
	format := localTime.Format("2006-01-02 15:04:05")
	nowZone, err := time.Parse("2006-01-02 15:04:05", format)
	if err != nil {
		fmt.Errorf("cant parse date: %#v", err)
	}*/
	// 截断到小时
	/*truncatedTime := nowZone.Truncate(time.Hour)
	fmt.Println("截断到小时:", truncatedTime)*/
	truncatedTime := nowZone
	// 获取时间的小时
	hour := truncatedTime.Hour()
	// 判断时间是否在晚上 11 点到次日 9 点之间
	if hour >= 22 && hour <= 24 {
		truncatedTime = time.Date(truncatedTime.Year(), truncatedTime.Month(), truncatedTime.Day(), 22, 0, 0, 0, truncatedTime.Location())
	} else if hour >= 0 && hour < 9 {
		truncatedTime = time.Date(truncatedTime.Year(), truncatedTime.Month(), truncatedTime.Day()-1, 22, 0, 0, 0, truncatedTime.Location())
	}

	return CustomTime{
		Time:      truncatedTime,
		IsNotZero: !truncatedTime.IsZero(),
	}
}

func GetSendClusteringTime(afterHour int, startTime CustomTime) CustomTime {
	// 获取当前时间后5小时的时间
	nowZone := startTime.Add(time.Duration(afterHour) * time.Hour)
	// 获取当前地区的时间
	//localTime := newTime.In(time.Local) // 使用本地时区
	//format := localTime.Format("2006-01-02 15:04:05")
	//nowZone, err := time.Parse("2006-01-02 15:04:05", format)
	//if err != nil {
	//	fmt.Errorf("cant parse date: %#v", err)
	//}
	// 截断到小时
	//truncatedTime := nowZone.Truncate(time.Hour)
	//fmt.Println("截断到小时:", truncatedTime)
	truncatedTime := nowZone
	// 获取时间的小时
	hour := truncatedTime.Hour()
	// 判断时间是否在晚上 11 点到次日 9 点之间
	if hour >= 22 && hour <= 24 {
		truncatedTime = time.Date(truncatedTime.Year(), truncatedTime.Month(), truncatedTime.Day()+1, 9, 0, 0, 0, truncatedTime.Location())
	} else if hour >= 0 && hour < 9 {
		truncatedTime = time.Date(truncatedTime.Year(), truncatedTime.Month(), truncatedTime.Day(), 9, 0, 0, 0, truncatedTime.Location())
	}

	return CustomTime{
		Time:      truncatedTime,
		IsNotZero: !truncatedTime.IsZero(),
	}
}

func CheckDiffTime(start CustomTime, nowTime CustomTime, minuteDiff int) bool {
	// 计算两个时间点之间的差值
	duration := nowTime.Sub(start.Time)
	// 将差值转换为分钟数
	minutesCha := int(duration.Minutes())
	if minutesCha-600 >= minuteDiff {
		return true
	} else {
		// 定义夜间时间范围
		nightStart := time.Date(start.Year(), start.Month(), start.Day(), 23, 0, 0, 0, start.Location())
		nightEnd := time.Date(start.Year(), start.Month(), start.Day()+1, 9, 0, 0, 0, start.Location())
		if nowTime.Before(nightStart) && start.Before(nightStart) {
			if minutesCha >= minuteDiff {
				return true
			} else {
				return false
			}
		} else if nowTime.After(nightEnd) && start.After(nightStart) {
			startToEndDiff := nightEnd.Sub(start.Time)
			startToEndDiffMinute := int(startToEndDiff.Minutes())
			if minutesCha-startToEndDiffMinute >= minuteDiff {
				return true
			} else {
				return false
			}
		} else {
			// nowTime.After(nightEnd) && start.Before(nightStart)
			return false
		}
	}
}

func (c CustomTime) IsNotDisturbTime() bool {
	// 获取时间的小时
	hour := c.Time.Hour()

	if config.ApplicationConfig.IsDebug {
		return false
	}
	// 判断时间是否在晚上 11 点到次日 9 点之间
	return hour >= 23 || hour < 9
}

func GetReportCountTime() (CustomTime, CustomTime) {
	startTime := GetNowCustomTime()
	//startTime, _ := GetCustomTimeByTime("1734118779")

	startReportTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day()-1, 0, 0, 0, 0, startTime.Location())
	startReportCustomTime := CustomTime{
		Time:      startReportTime,
		IsNotZero: !startReportTime.IsZero(),
	}

	endReportTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	endReportCustomTime := CustomTime{
		Time:      endReportTime,
		IsNotZero: !endReportTime.IsZero(),
	}
	return startReportCustomTime, endReportCustomTime
}
