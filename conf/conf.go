package conf

import (
	"path/filepath"

	"github.com/fufuok/cmap"
	"github.com/fufuok/utils"
	"github.com/imroc/req"
)

var (
	Debug     bool
	Version   = "v0.0.0"
	GoVersion = ""
	GitCommit = ""

	// RootPath 运行绝对路径
	RootPath = utils.ExecutableDir(true)

	// FilePath 配置文件绝对路径
	FilePath = filepath.Join(RootPath, "..", "etc")

	// ConfigFile 默认配置文件路径
	ConfigFile = filepath.Join(FilePath, ProjectName+".json")

	// LogDir 日志路径
	LogDir  = filepath.Join(RootPath, "..", "log")
	LogFile = filepath.Join(LogDir, ProjectName+".log")

	// LogDaemon 守护日志
	LogDaemon = filepath.Join(LogDir, "daemon.log")

	// Config 所有配置
	Config tJSONConf

	// ReqUserAgent 请求名称
	ReqUserAgent = req.Header{"User-Agent": APPName + "/" + Version}

	// ESIPs 采集目标 IP 对应的 ES 源 IP 数据
	ESIPs = cmap.New()
)
