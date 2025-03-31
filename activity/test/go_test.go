package test

import (
	"fmt"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"log"
	"testing"
)

func TestGoContext(t *testing.T) {
	someFunction()
}

func TestGoContext2(t *testing.T) {
	str := "[{\"id\":\"1\",\"bodyText\":\"Hi，拜托帮我点一下助力，参加 magic chess 组队预约活动，可以获得限定英雄皮肤，抽千元现金大奖.\\\\n帮忙点一下，大奖就在眼前\\\\n{{1}}\",\"weight\":33},{\"id\":\"2\",\"bodyText\":\"Hi，2222，参加 magic chess 组队预约活动，可以获得限定英雄皮肤，抽千元现金大奖.\\\\n帮忙点一下，大奖就在眼前\\\\n{{1}}\",\"weight\":33},{\"id\":\"3\",\"bodyText\":\"Hi，3333，参加 magic chess 组队预约活动，可以获得限定英雄皮肤，抽千元现金大奖.\\\\n帮忙点一下，大奖就在眼前\\\\n{{1}}\",\"weight\":34}]"
	var helpTextList []config.HelpText
	err := json.NewEncoder().Decode([]byte(str), &helpTextList)
	if err != nil {
		log.Fatal(err)
	}
}

func someFunction() {
	defer func() {
		if r := recover(); r != nil {
			// 处理发生空指针错误的情况
			fmt.Println("Recovered from nil pointer dereference:", r)
		}
	}()

	// 可能引发空指针错误的代码
	var ptr *int
	_ = *ptr // 这里会触发空指针错误
}

func TestGoContext3(t *testing.T) {
	timestamp := "1737601995"
	msgRecTime, _ := util.GetCustomTimeByTime(timestamp)
	endTime := util.GetSendClusteringTime(3, msgRecTime)
	sendRenewFreeAt := util.GetSendRenewMsgTime(1, msgRecTime)
	log.Println(endTime)
	log.Println(sendRenewFreeAt)
}

func TestGoContext4(t *testing.T) {
	number := 2413443
	formattedNumber := util.AddThousandSeparators64(int64(number))
	fmt.Println(formattedNumber)
}

func TestGoCTime(t *testing.T) {
	number := 2413443
	formattedNumber := util.AddThousandSeparators64(int64(number))
	fmt.Println(formattedNumber)
}
