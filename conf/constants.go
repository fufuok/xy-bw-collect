package conf

import (
	"time"
)

const (
	APPName     = "XY.BWCollect"
	ProjectName = "xybwcollect"

	TrapCommunityName = "BW_TRAP_COMMUNITY"
	TrapServerAddr    = ":162"

	// TrapCacheTable Trap 报警信息缓存表名, 每秒一个
	TrapCacheTable = "TRAP:15:04:05"
	TrapAlarmCode  = "monitor_trap_alarm"

	SNMPCommunityName  = "BW_SNMP_COMMUNITY"
	SNMPPort           = "161"
	SNMPTimeout        = 10
	SNMPRetries        = 2
	SNMPMaxRepetitions = 10

	// DiscardDurationMin 丢弃花费时间较长的数据 (秒)
	DiscardDurationMin = 30
	// DiscardIntervalMin 丢弃间隔时间较短的数据 (秒)
	DiscardIntervalMin = 30
	// DiscardIntervalMax 丢弃间隔时间较长的数据 (秒)
	DiscardIntervalMax = 90

	// BaseSecretKeyName 项目基础密钥 (环境变量名)
	BaseSecretKeyName = "BW_BASE_SECRET_KEY"
	// BaseSecretSalt 用于解密基础密钥值的密钥 (编译在程序中)
	BaseSecretSalt = "Fufu^Bw .Ks"

	// ESBodySep ES 数据分隔符
	ESBodySep = "=-:-="

	// ESPostBatchNum ES 单次批量写入最大条数或最大字节数
	ESPostBatchNum   = 3000
	ESPOSTBatchBytes = 30 << 20

	// LogLevel 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	LogLevel = 3
	// LogCacheTable Log 缓存表名, 每分钟一个
	LogCacheTable = "LOG:15:04"
	// LogSamplePeriodDur 抽样日志设置 (每秒最多 3 个日志)
	LogSamplePeriodDur = time.Second
	LogSampleBurst     = 3
	// LogFileMaxSize 每 100M 自动切割, 保留 30 天内最近 10 个日志文件
	LogFileMaxSize    = 100
	LogFileMaxBackups = 10
	LogFileMaxAge     = 30

	// BWCacheTable 采集数据缓存表名, 每分钟一个
	BWCacheTable = "BW:15:04"

	// WatcherInterval 文件变化监控时间间隔(分)
	WatcherInterval = 1
)
