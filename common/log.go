package common

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fufuok/cache2go"
	"github.com/fufuok/utils/xid"
	"github.com/imroc/req"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/fufuok/xy-bw-collect/conf"
)

var (
	Log        zerolog.Logger
	LogSampled zerolog.Logger
)

type ESWriter struct{}

func NewESWriter() *ESWriter {
	return &ESWriter{}
}

func (w *ESWriter) Write(p []byte) (n int, err error) {
	go LogCache(append([]byte{}, p...))
	return len(p), nil
}

func initLogger() {
	if err := InitLogger(); err != nil {
		log.Fatalln("Failed to initialize logger:", err, "\nbye.")
	}

	// 路径脱敏, 日志格式规范, 避免与自定义字段名冲突: {"E":"is Err(error)","error":"is Str(error)"}
	zerolog.TimestampFieldName = "T"
	zerolog.LevelFieldName = "L"
	zerolog.MessageFieldName = "M"
	zerolog.ErrorFieldName = "E"
	zerolog.CallerFieldName = "F"
	zerolog.ErrorStackFieldName = "S"
	zerolog.DurationFieldInteger = true
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
}

func InitLogger() error {
	if err := LogConfig(); err != nil {
		return err
	}

	// 抽样的日志记录器
	sampler := &zerolog.BurstSampler{
		Burst:  conf.Config.SYSConf.Log.Burst,
		Period: conf.Config.SYSConf.Log.PeriodDur,
	}
	LogSampled = Log.Sample(&zerolog.LevelSampler{
		TraceSampler: sampler,
		DebugSampler: sampler,
		InfoSampler:  sampler,
		WarnSampler:  sampler,
		ErrorSampler: sampler,
	})

	req.Debug = conf.Debug

	return nil
}

// LogConfig 日志配置
// 1. Debug 时, 高亮输出到控制台
// 2. 生产环境时, 输出到日志文件(可选关闭高亮, 保存最近 10 个 30 天内的日志), 并发送 JSON 日志到 ES
func LogConfig() error {
	var (
		writers  []io.Writer
		basicLog = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "0102 15:04:05"}
	)

	if conf.Debug {
		writers = []io.Writer{basicLog}
	} else {
		basicLog.NoColor = conf.Config.SYSConf.Log.NoColor
		basicLog.Out = &lumberjack.Logger{
			Filename:   conf.Config.SYSConf.Log.File,
			MaxSize:    conf.Config.SYSConf.Log.MaxSize,
			MaxAge:     conf.Config.SYSConf.Log.MaxAge,
			MaxBackups: conf.Config.SYSConf.Log.MaxBackups,
			LocalTime:  true,
			Compress:   true,
		}
		writers = []io.Writer{basicLog, NewESWriter()}
	}

	Log = zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Caller().Logger()
	Log = Log.Level(zerolog.Level(conf.Config.SYSConf.Log.Level))

	return nil
}

// LogCache 日志暂存
func LogCache(bs []byte) {
	key := xid.NewString()
	table := time.Now().Format(conf.LogCacheTable)
	cache := cache2go.Cache(table)
	cache.Add(key, 180*time.Second, bs)
}
