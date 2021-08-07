package main

import (
	"github.com/fufuok/utils/xdaemon"

	"github.com/fufuok/xy-bw-collect/conf"
	"github.com/fufuok/xy-bw-collect/master"
)

func main() {
	if !conf.Config.SYSConf.Debug {
		xdaemon.NewDaemon(conf.LogDaemon).Run()
	}

	master.Start()
	master.Watcher()

	select {}
}
