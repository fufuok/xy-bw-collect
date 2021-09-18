package master

import (
	"context"
	"os"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/service"
)

var (
	// 重启信号
	restartChan = make(chan bool)

	// 配置重载信息
	reloadChan = make(chan bool)
)

func Start() {
	go func() {
		// 初始化
		initMaster()

		for {
			// 获取远程配置
			ctx, cancel := context.WithCancel(context.Background())
			go startRemoteConf(ctx)

			// 采集器配置更新
			go service.InitCollector()

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

func initMaster() {
	// 优先初始化公共变量
	common.InitCommon()

	// 启动数据服务
	service.InitService()
}
