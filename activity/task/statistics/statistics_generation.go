package statistics

import (
	"context"
	"errors"
	"fmt"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
)

func GenerationInfo(ctx context.Context, timeRange dto.StatisticsTimeRange) ([]dto.GenerationUserDto, error) {
	limit := 100
	minId := 0
	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	channelMap := make(map[string]map[string]map[string]*dto.GenerationUserDto)
	/*初始化数量统计map*/
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		channelMap[channel] = make(map[string]map[string]*dto.GenerationUserDto)
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			channelMap[channel][language] = make(map[string]*dto.GenerationUserDto)
		}
	}
	for {
		generationInfoList, err := userAttendInfoMapper.SelectListByGeneration(timeRange.StartTimestamp, timeRange.EndTimestamp, minId, limit)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询userAttendInfo失败,err：%v", "GenerationInfo", err))
			return nil, errors.New("database is error")
		}
		if len(generationInfoList) == 0 {
			break
		}
		for _, generationInfo := range generationInfoList {
			if minId < generationInfo.Id {
				minId = generationInfo.Id
			}
			generationMap := channelMap[generationInfo.Channel][generationInfo.Language]
			if generationMap == nil {
				generationMap = make(map[string]*dto.GenerationUserDto)
			}
			if generationMap[generationInfo.Generation] == nil {
				generationMap[generationInfo.Generation] = &dto.GenerationUserDto{
					Generation: generationInfo.Generation,
					Language:   generationInfo.Language,
					Channel:    generationInfo.Channel,
					Count:      0,
				}
			}
			generationMap[generationInfo.Generation].Count += 1
		}
	}
	generationUserDtoList := make([]dto.GenerationUserDto, 0)

	for _, languageMap := range channelMap {
		for _, generationMap := range languageMap {
			for _, generationInfo := range generationMap {
				generationUserDto := dto.GenerationUserDto{
					Generation: generationInfo.Generation,
					Count:      generationInfo.Count,
					Channel:    generationInfo.Channel,
					Language:   generationInfo.Language,
				}
				generationUserDtoList = append(generationUserDtoList, generationUserDto)
			}
		}
	}
	return generationUserDtoList, nil
}

func GenerationInfoWithAttend(ctx context.Context, timeRange dto.StatisticsTimeRange) ([]dto.GenerationUserDto, error) {
	limit := 100
	minId := 0
	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	attendMap := make(map[string]map[string]*dto.GenerationUserDto)
	attendMap[constant.AttendStatusAttend] = make(map[string]*dto.GenerationUserDto)
	attendMap[constant.AttendStatusStartGroup] = make(map[string]*dto.GenerationUserDto)
	attendMap[constant.AttendStatusEightOver] = make(map[string]*dto.GenerationUserDto)
	for {
		generationInfoList, err := userAttendInfoMapper.SelectListByGeneration(timeRange.StartTimestamp, timeRange.EndTimestamp, minId, limit)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询userAttendInfo失败,err：%v", "GenerationInfo", err))
			return nil, errors.New("database is error")
		}
		if len(generationInfoList) == 0 {
			break
		}
		for _, generationInfo := range generationInfoList {
			if minId < generationInfo.Id {
				minId = generationInfo.Id
			}
			generationMap := attendMap[generationInfo.AttendStatus]
			if generationMap == nil {
				generationMap = make(map[string]*dto.GenerationUserDto)
			}
			if generationMap[generationInfo.Generation] == nil {
				generationMap[generationInfo.Generation] = &dto.GenerationUserDto{
					Generation:   generationInfo.Generation,
					AttendStatus: generationInfo.AttendStatus,
					Count:        0,
				}
			}
			generationMap[generationInfo.Generation].Count += 1
		}
	}
	generationUserDtoList := make([]dto.GenerationUserDto, 0)
	for _, generationMap := range attendMap {
		for _, generationInfo := range generationMap {
			generationUserDto := dto.GenerationUserDto{
				Generation:   generationInfo.Generation,
				Count:        generationInfo.Count,
				AttendStatus: generationInfo.AttendStatus,
			}
			generationUserDtoList = append(generationUserDtoList, generationUserDto)
		}
	}
	return generationUserDtoList, nil
}
