package master

import (
	"github.com/fufuok/xy-bw-collect/service"
)

// 启动采集器
func startCollector() {
	go service.InitCollector()
}
