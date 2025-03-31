package statistics

import (
	"context"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/config"
	"strconv"
)

func MsgInfo(ctx context.Context, timeRange dto.StatisticsTimeRange) (map[string][]dto.ReFreeCountDto, error) {
	limit := 100
	minId := int64(0)
	msgTypes := []string{constant.RenewFreeMsg, constant.PayRenewFreeMsg, constant.PromoteClusteringMsg}
	msgInfoMapper := dao.GetMsgInfoMapperV2()
	userAttendMapper := dao.GetUserAttendInfoMapperV2()
	msgCountMap := make(map[string]map[string]map[string]*dto.ReFreeCountDto)
	/*初始化数量统计map*/
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		msgCountMap[channel] = make(map[string]map[string]*dto.ReFreeCountDto)
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			msgCountMap[channel][language] = make(map[string]*dto.ReFreeCountDto)
			for _, msgType := range msgTypes {
				msgCountMap[channel][language][msgType] = &dto.ReFreeCountDto{Channel: channel, Language: language, MsgType: msgType}
			}
		}
	}
	/*循环查表统计数量*/
	for {
		msgInfos, _ := msgInfoMapper.SelectListByMsgType(timeRange.StartTimestamp, timeRange.EndTimestamp, strconv.FormatInt(minId, 10), limit)
		count := len(msgInfos)
		if count == 0 {
			break
		}
		waIds := getDistinctMsgInfoWaId(msgInfos)
		userAttendInfoEntityList, _ := userAttendMapper.SelectListByWaIds(waIds)
		if len(userAttendInfoEntityList) == 0 {
			break
		}
		userAttendInfoMap := make(map[string]*entity.UserAttendInfoEntityV2)
		for _, userAttendInfo := range userAttendInfoEntityList {
			userAttendInfoMap[userAttendInfo.WaId] = &userAttendInfo
		}
		for _, msgInfo := range msgInfos {
			waId := msgInfo.WaId
			userAttendInfo := userAttendInfoMap[waId]
			if userAttendInfo == nil {
				continue
			}
			channel := userAttendInfo.Channel
			language := userAttendInfo.Language
			msgType := msgInfo.MsgType
			if msgCountMap[channel][language][msgType] != nil {
				msgCountMap[channel][language][msgType].Count++
			}
			msgInfoId, _ := strconv.ParseInt(msgInfo.Id, 10, 64)
			if minId < msgInfoId {
				minId = msgInfoId
			}
		}
	}
	/*整理返回结构*/
	result := make(map[string][]dto.ReFreeCountDto)
	for _, msgType := range msgTypes {
		result[msgType] = make([]dto.ReFreeCountDto, 0)
		for _, channel := range config.ApplicationConfig.Activity.ChannelList {
			for _, language := range config.ApplicationConfig.Activity.LanguageList {
				result[msgType] = append(result[msgType], *msgCountMap[channel][language][msgType])
			}
		}
	}
	return result, nil
}
func getDistinctMsgInfoWaId(s []entity.MsgInfoEntityV2) []string {
	result := make([]string, 0)
	m := make(map[string]bool) //map的值不重要
	for _, v := range s {
		if _, ok := m[v.WaId]; !ok {
			result = append(result, v.WaId)
			m[v.WaId] = true
		}
	}
	return result
}
