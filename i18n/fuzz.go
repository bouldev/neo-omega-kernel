package i18n

import (
	"fmt"
	"strings"
)

func CheckAndAppend(ifHas string, addDescribe string) func(error) error {
	return func(err error) error {
		if strings.Contains(err.Error(), ifHas) {
			return fmt.Errorf("\n%v\n解释:%v", err.Error(), addDescribe)
		}
		return nil
	}
}

func FuzzyTransErr(err error) error {
	rules := []func(error) error{
		CheckAndAppend("EOF", "文件或者网络连接(过早)中断,如果没有更多有价值信息,可能无法推断错误真正原因,请尝试联系开发者"),
		CheckAndAppend("kick", "机器人被从服务器中踢出去了,可能是管理做的或者机器人自己被谁错误封禁了(请删除封禁数据)"),
		CheckAndAppend("loggedinOtherLocation", "多个程序在使用同一个机器人,请检查是否有多个程序在同一个服务器作业"),
		CheckAndAppend("Already logged in", "机器人已经登陆 -> 可能是机器人刚刚退出就重进，过几秒再进就好"),
		CheckAndAppend("error code: 522", "验证服务器已宕机"),
	}
	for _, r := range rules {
		if newErr := r(err); newErr != nil {
			return newErr
		}
	}
	return err
}
