package txUtil

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhuxiujia/GoMybatis"
	"github.com/zhuxiujia/GoMybatis/tx"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
)

const transactionKey string = "transactionSession"

// 开始一个新的事务并将其放入上下文中
func startTransaction(ctx *gin.Context) (GoMybatis.Session, error) {
	session, err := dao.GetCostCountInfoMapper().SessionSupport.NewSession()
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
		return nil, errors.New("database is error")
	}
	p := tx.NewPropagation("")
	session.Begin(&p)
	ctx.Set(transactionKey, session)
	// 将事务放入上下文中
	return session, nil
}

// GetTransaction 从上下文中获取事务
func GetTransaction(ctx *gin.Context) (GoMybatis.Session, bool, error) {
	session, ok := ctx.Get(transactionKey)
	if !ok {
		session, err := startTransaction(ctx)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
			return nil, false, errors.New("database is error")
		}
		return session, ok, nil
	}
	return session.(GoMybatis.Session), ok, nil
}
