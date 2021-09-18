package service

import (
	"github.com/fufuok/xy-bw-collect/service/snmpv2c"
)

func InitService() {
	// 启动 Trap 服务
	go snmpv2c.InitTrapServer()

	// 心跳服务
	go initHeartbeat()

	// 采集数据上报
	go initReportBW()

	// 发送 Trap 报警
	go initReportTrap()

	// 初始化运行时参数
	go initRuntime()
}
