package service

import (
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
	"github.com/fufuok/xy-bw-collect/service/snmpv2c"
)

// InitCollector 初始化新增或变动的配置
func InitCollector() {
	if conf.Config.SYSConf.Debug {
		common.Log.Info().Msgf("init new collector: %s", conf.Config.SNMPConf.V2.TargetNew)
	}
	for targetKey := range conf.Config.SNMPConf.V2.TargetNew {
		go runCollector(targetKey)
	}
}

// 执行数据采集
func runCollector(targetKey string) {
	utils.WaitNextMinute()
	info := common.NewIFInfo()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// 每分钟采集
	for range ticker.C {
		target, ok := conf.Config.SNMPConf.V2.Target[targetKey]
		if !ok {
			// 配置已变更, 退出采集
			common.Log.Info().Str("target", targetKey).Msg("exit collect")
			break
		}
		snmpv2c.Collector(target, info)
	}
}
