package master

import (
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
	"github.com/fufuok/xy-bw-collect/service"
)

// 每分钟写一次日志
func startLogAPI() {
	for range time.Tick(time.Minute) {
		if conf.Config.SYSConf.LogAPI != "" {
			// 心跳日志
			common.LogCache(utils.MustJSON(map[string]interface{}{
				"type":          "heartbeat",
				"internal_ipv4": service.InternalIPv4,
				"external_ipv4": service.ExternalIPv4,
				"time":          time.Now().Format(time.RFC3339),
			}))
			go common.SendLastData(-time.Minute, conf.LogCacheTable, conf.Config.SYSConf.LogAPI)
		}
	}
}

// 每分钟上报采集数据
func startReportBW() {
	for range time.Tick(time.Minute) {
		go common.SendLastData(-time.Minute, conf.BWCacheTable, conf.Config.SYSConf.ReportAPI)
	}
}

// 每秒发送 Trap 报警数据
func startReportTrap() {
	for range time.Tick(time.Second) {
		go common.SendLastData(-time.Second, conf.TrapCacheTable, conf.Config.TrapConf.ReportAPI)
	}
}
