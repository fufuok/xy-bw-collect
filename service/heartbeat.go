package service

import (
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
)

// 心跳日志
func initHeartbeat() {
	for range time.Tick(time.Minute) {
		if conf.Config.SYSConf.LogAPI != "" {
			common.LogCache(utils.MustJSON(map[string]interface{}{
				"type":          "heartbeat",
				"internal_ipv4": InternalIPv4,
				"external_ipv4": ExternalIPv4,
				"time":          time.Now().Format(time.RFC3339),
			}))
			go common.SendLastData(-time.Minute, conf.LogCacheTable, conf.Config.SYSConf.LogAPI)
		}
	}
}
