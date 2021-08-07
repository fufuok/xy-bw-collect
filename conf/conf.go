package conf

import (
	"path/filepath"

	"github.com/fufuok/cmap"
	"github.com/fufuok/utils"
	"github.com/imroc/req"
)

// RootPath 运行绝对路径
var RootPath = utils.ExecutableDir(true)

// FilePath 配置文件绝对路径
var FilePath = filepath.Join(RootPath, "..", "etc")

// ConfigFile 默认配置文件路径
var ConfigFile = filepath.Join(FilePath, ProjectName+".json")

// LogDir 日志路径
var LogDir = filepath.Join(RootPath, "..", "log")
var LogFile = filepath.Join(LogDir, ProjectName+".log")

// LogDaemon 守护日志
var LogDaemon = filepath.Join(LogDir, "daemon.log")

// Config 所有配置
var Config tJSONConf

// ReqUserAgent 请求名称
var ReqUserAgent = req.Header{"User-Agent": APPName + "/" + CurrentVersion}

// ESIPs 采集目标 IP 对应的 ES 源 IP 数据
var ESIPs = cmap.New()
