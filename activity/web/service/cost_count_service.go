package service

import (
	"context"
	"errors"
	"fmt"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
)

func CostIsUltraLimit(ctx context.Context) (bool, error) {
	methodName := "CostIsUltraLimit"
	costCountMapper := dao.GetCostCountInfoMapper()
	costCount, err := costCountMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取已使用费用失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return true, err
	}
	// 查询活动信息
	activityInfoMapper := dao.GetActivityInfoMapper()
	activityInfo, err := activityInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return true, err
	}
	if activityInfo.Id <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息不存在，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return true, errors.New(fmt.Sprintf("方法[%s],查询活动信息不存在，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	}
	if costCount.Id > 0 && activityInfo.CostMax <= costCount.CostCount {
		return true, nil
	}
	return false, nil
}
