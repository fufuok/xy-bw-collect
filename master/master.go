package master

import (
	"context"
	"os"

	"github.com/fufuok/xy-bw-collect/common"
)

var (
	// 重启信号
	restartChan = make(chan bool)
	// 配置重载信息
	reloadChan = make(chan bool)
)

func Start() {
	go func() {
		go startTrapServer()

		go startLogAPI()
		go startReportBW()
		go startReportTrap()

		for {
			// 获取远程配置
			ctx, cancel := context.WithCancel(context.Background())
			go startRemoteConf(ctx)

			// 采集器
			go startCollector()

			select {
			case <-restartChan:
				// 强制退出, 由 Daemon 重启程序
				common.Log.Warn().Msg("restart <-restartChan")
				os.Exit(0)
			case <-reloadChan:
				cancel()
				common.Log.Warn().Msg("reload <-reloadChan")
			}
		}
	}()
}
