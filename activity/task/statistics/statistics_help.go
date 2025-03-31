package statistics

import (
	"context"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/config"
)

func HelpInfo(ctx context.Context, timeRange dto.StatisticsTimeRange) ([]dto.HelpCountDto, error) {
	limit := 100
	minId := "0"
	helpInfoMapper := dao.GetHelpInfoMapperV2()
	userAttendMapper := dao.GetUserAttendInfoMapperV2()
	creatorInfoHelpCountMap := make(map[string]map[string]map[int]*dto.HelpCountDto)
	/*初始化数量统计map*/
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		creatorInfoHelpCountMap[channel] = make(map[string]map[int]*dto.HelpCountDto)
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			creatorInfoHelpCountMap[channel][language] = make(map[int]*dto.HelpCountDto)
			for i := 1; i < 9; i++ {
				creatorInfoHelpCountMap[channel][language][i] = &dto.HelpCountDto{Channel: channel, Language: language, HelpNum: i}
			}
		}
	}
	/*循环查表统计数量*/
	for {
		helpCaches, err := helpInfoMapper.SelectDistinctCodeByTimestamp(timeRange.StartTimestamp, timeRange.EndTimestamp, minId, limit)
		if err != nil {
			return nil, err
		}
		codes := make([]string, 0)
		for _, helpCache := range helpCaches {
			if helpCache.RallyCode == "" {
				continue
			}
			minId = helpCache.RallyCode
			codes = append(codes, helpCache.RallyCode)
		}
		count := len(codes)
		if count == 0 {
			break
		}
		creatorHelpedCounts, _ := helpInfoMapper.CountByCodesTimestamp(codes, timeRange.StartTimestamp, timeRange.EndTimestamp)
		creatorHelpCountMap := make(map[string]int, len(creatorHelpedCounts))
		for _, creatorHelpCount := range creatorHelpedCounts {
			creatorHelpCountMap[creatorHelpCount.Code] = creatorHelpCount.Count
		}
		userAttendInfoEntityList, _ := userAttendMapper.SelectListByCodes(codes)
		for _, entity := range userAttendInfoEntityList {
			helpCount := creatorHelpCountMap[entity.RallyCode]
			if helpCount > 8 {
				helpCount = 8
			}
			if creatorInfoHelpCountMap[entity.Channel] == nil {
				continue
			}
			if creatorInfoHelpCountMap[entity.Channel][entity.Language] == nil {
				continue
			}
			if creatorInfoHelpCountMap[entity.Channel][entity.Language][helpCount] == nil {
				continue
			}
			creatorInfoHelpCountMap[entity.Channel][entity.Language][helpCount].HelpNumCount++
		}
	}
	/*重构返回数据结构*/
	result := make([]dto.HelpCountDto, 0)
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			for i := 1; i < 9; i++ {
				helpCountDto := creatorInfoHelpCountMap[channel][language][i]
				result = append(result, *helpCountDto)
			}
		}
	}
	return result, nil
}
