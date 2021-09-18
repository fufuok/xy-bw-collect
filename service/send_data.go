package service

import (
	"time"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
)

// 每分钟上报采集数据
func initReportBW() {
	for range time.Tick(time.Minute) {
		go common.SendLastData(-time.Minute, conf.BWCacheTable, conf.Config.SYSConf.ReportAPI)
	}
}

// 每秒发送 Trap 报警数据
func initReportTrap() {
	for range time.Tick(time.Second) {
		go common.SendLastData(-time.Second, conf.TrapCacheTable, conf.Config.TrapConf.ReportAPI)
	}
}
