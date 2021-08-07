package master

import (
	"github.com/fufuok/xy-bw-collect/service/snmpv2c"
)

// 启动 Trap 服务
func startTrapServer() {
	go snmpv2c.InitTrapServer()
}
