package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/internal/json"
)

// 接口配置
type tJSONConf struct {
	SYSConf  tSYSConf  `json:"sys_conf"`
	SNMPConf tSNMPConf `json:"snmp_conf"`
	TrapConf tTrapConf `json:"trap_conf"`
}

type tSYSConf struct {
	Debug           bool       `json:"debug"`
	Log             tLogConf   `json:"log"`
	MainConfig      TFilesConf `json:"main_config"`
	RestartMain     bool       `json:"restart_main"`
	WatcherInterval int        `json:"watcher_interval"`
	LogAPI          string     `json:"log_api"`
	ReportAPI       string     `json:"report_api"`
	BaseSecretValue string
}

type tLogConf struct {
	Level      int    `json:"level"`
	NoColor    bool   `json:"no_color"`
	File       string `json:"file"`
	Period     int    `json:"period"`
	Burst      uint32 `json:"burst"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	PeriodDur  time.Duration
}

type tSNMPConf struct {
	CommunityName      string  `json:"community_name"`
	Port               int     `json:"port"`
	BanRegexpValue     string  `json:"ban_regexp"`
	Timeout            int     `json:"timeout"`
	Retries            uint    `json:"retries"`
	MaxRepetitions     int     `json:"max_repetitions"`
	DiscardDurationMin int     `json:"discard_duration_min"`
	DiscardIntervalMin float64 `json:"discard_interval_min"`
	DiscardIntervalMax float64 `json:"discard_interval_max"`
	V2                 tV2Conf `json:"v2"`
	Community          string
	BanRegexp          *regexp.Regexp `json:"-"`
	DiscardDuration    time.Duration
}

type tV2Conf struct {
	Files     []TFilesConf `json:"files"`
	AddrConf  []string     `json:"addr"`
	ConfFile  []string
	Target    map[string]TV2Target
	TargetNew map[string]struct{}
}

type TV2Target struct {
	Community string
	Addr      string
	Port      string
}

// TESIPs ES snmpip 中获取的源 IP 数据, 上报数据时附带上报
type TESIPs struct {
	NodeIP string `json:"node_ip"`
	SNMPIP string `json:"snmp_ip"`
}

type TFilesConf struct {
	Path            string      `json:"path"`
	Method          string      `json:"method"`
	SecretName      string      `json:"secret_name"`
	API             string      `json:"api"`
	ESIndex         string      `json:"es_index"`
	ESBody          interface{} `json:"es_body"`
	Interval        int         `json:"interval"`
	SecretValue     string
	GetConfDuration time.Duration
	ConfigMD5       string
	ConfigVer       time.Time
}

// tTrapConf trap server 配置项, 接收的 Oids, Identifiers
type tTrapConf struct {
	ReportAPI     string `json:"report_api"`
	AlarmCode     string `json:"alarm_code"`
	LocalAddr     string `json:"local_addr"`
	CommunityName string `json:"community_name"`

	// 标识符描述和 Oid 对照表: 1.3.6.1.6.3.1.1.5.3: Down
	Identifiers map[string]string `json:"identifiers"`

	// 限定收集的 Oids 前缀: 1.3.6.1.2.1.2.2.1.2
	InterfacePrefix []string `json:"interface_prefix"`

	// 报警 IP 和名称对照表
	IPNameTable map[string]string `json:"ip_name_table"`

	InterfacePrefixOids snmpgo.Oids

	Community string
}

func init() {
	confFile := flag.String("c", ConfigFile, "配置文件绝对路径")
	flag.Parse()
	ConfigFile = *confFile
	if err := LoadConf(); err != nil {
		log.Fatalln("Failed to initialize config:", err, "\nbye.")
	}
}

// LoadConf 加载配置
func LoadConf() error {
	config, err := readConf()
	if err != nil {
		return err
	}

	Config = *config

	return nil
}

// 读取配置
func readConf() (*tJSONConf, error) {
	body, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	config := new(tJSONConf)
	if err := json.Unmarshal(body, config); err != nil {
		return nil, err
	}

	// 基础密钥 Key
	config.SYSConf.BaseSecretValue = utils.GetenvDecrypt(BaseSecretKeyName, BaseSecretSalt)
	if config.SYSConf.BaseSecretValue == "" {
		return nil, fmt.Errorf("%s cannot be empty", BaseSecretKeyName)
	}

	// 解密 SNMP Community, 配置中的环境变量名优先
	envName := config.SNMPConf.CommunityName
	if envName == "" {
		envName = SNMPCommunityName
	}
	config.SNMPConf.Community = utils.GetenvDecrypt(envName, config.SYSConf.BaseSecretValue)
	if config.SNMPConf.Community == "" {
		return nil, fmt.Errorf("%s cannot be empty", envName)
	}

	// 整理目标和密码
	v2Target := make(map[string]TV2Target)
	v2TargetNew := make(map[string]struct{})

	// 整合配置文件设备列表
	for i, f := range config.SNMPConf.V2.Files {
		file := strings.TrimSpace(f.Path)
		if file == "" || strings.HasPrefix(file, "#") {
			config.SNMPConf.V2.Files[i].Path = ""
			continue
		}

		// 解密 SecretName, 各类 API 的加密参数值
		if f.SecretName != "" {
			config.SNMPConf.V2.Files[i].SecretValue = utils.GetenvDecrypt(f.SecretName,
				config.SYSConf.BaseSecretValue)
			if config.SNMPConf.V2.Files[i].SecretValue == "" {
				return nil, fmt.Errorf("%s cannot be empty", f.SecretName)
			}
		}

		// 补全文件路径
		if !filepath.IsAbs(file) {
			file = filepath.Join(FilePath, file)
		}

		body, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if err := parseAddrConf(strings.Split(string(body), "\n"),
			config.SNMPConf.Community, v2Target, v2TargetNew); err != nil {
			return nil, err
		}

		// 每次获取配置的时间间隔
		if f.Interval > 29 {
			config.SNMPConf.V2.Files[i].GetConfDuration = time.Duration(f.Interval) * time.Second
		}

		config.SNMPConf.V2.Files[i].Path = file
	}

	// 整合指定的设备列表
	if err := parseAddrConf(config.SNMPConf.V2.AddrConf,
		config.SNMPConf.Community, v2Target, v2TargetNew); err != nil {
		return nil, err
	}

	config.SNMPConf.V2.Target = v2Target
	config.SNMPConf.V2.TargetNew = v2TargetNew

	// 连接超时, 最大重试次数, 最大重复次数
	if config.SNMPConf.Timeout < 1 {
		config.SNMPConf.Timeout = SNMPTimeout
	}
	if config.SNMPConf.Retries < 0 {
		config.SNMPConf.Retries = SNMPRetries
	}
	if config.SNMPConf.MaxRepetitions < 1 {
		config.SNMPConf.MaxRepetitions = SNMPMaxRepetitions
	}

	// 丢弃花费时间较长的数据
	if config.SNMPConf.DiscardDurationMin < 1 {
		config.SNMPConf.DiscardDurationMin = DiscardDurationMin
	}
	config.SNMPConf.DiscardDuration = time.Duration(config.SNMPConf.DiscardDurationMin) * time.Second

	// 丢弃间隔时间较短的数据
	if config.SNMPConf.DiscardIntervalMin < 1 {
		config.SNMPConf.DiscardIntervalMin = DiscardIntervalMin
	}

	// 丢弃间隔时间较长的数据
	if config.SNMPConf.DiscardIntervalMax < 1 {
		config.SNMPConf.DiscardIntervalMax = DiscardIntervalMax
	}

	// 接口名称黑名单
	if config.SNMPConf.BanRegexpValue != "" {
		banRegexp, err := regexp.Compile(config.SNMPConf.BanRegexpValue)
		if err != nil {
			return nil, err
		}
		config.SNMPConf.BanRegexp = banRegexp
	}

	// 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	if config.SYSConf.Log.Level > 7 || config.SYSConf.Log.Level < -1 {
		config.SYSConf.Log.Level = LogLevel
	}

	// 抽样日志设置 (x 秒 n 条)
	if config.SYSConf.Log.Burst < 0 || config.SYSConf.Log.Period < 0 {
		config.SYSConf.Log.PeriodDur = LogSamplePeriodDur
		config.SYSConf.Log.Burst = LogSampleBurst
	} else {
		config.SYSConf.Log.PeriodDur = time.Duration(config.SYSConf.Log.Period) * time.Second
	}

	// 日志文件
	if config.SYSConf.Log.File == "" {
		config.SYSConf.Log.File = LogFile
	}

	// 日志大小和保存设置
	if config.SYSConf.Log.MaxSize < 1 {
		config.SYSConf.Log.MaxSize = LogFileMaxSize
	}
	if config.SYSConf.Log.MaxBackups < 1 {
		config.SYSConf.Log.MaxBackups = LogFileMaxBackups
	}
	if config.SYSConf.Log.MaxAge < 1 {
		config.SYSConf.Log.MaxAge = LogFileMaxAge
	}

	// 每次获取远程主配置的时间间隔, < 30 秒则禁用该功能
	if config.SYSConf.MainConfig.Interval > 29 {
		// 远程获取主配置 API, 解密 SecretName
		if config.SYSConf.MainConfig.SecretName != "" {
			config.SYSConf.MainConfig.SecretValue = utils.GetenvDecrypt(config.SYSConf.MainConfig.SecretName,
				config.SYSConf.BaseSecretValue)
			if config.SYSConf.MainConfig.SecretValue == "" {
				return nil, fmt.Errorf("%s cannot be empty", config.SYSConf.MainConfig.SecretName)
			}
		}
		config.SYSConf.MainConfig.GetConfDuration = time.Duration(config.SYSConf.MainConfig.Interval) * time.Second
		config.SYSConf.MainConfig.Path = ConfigFile
	}

	// 文件变化监控时间间隔
	if config.SYSConf.WatcherInterval < 1 {
		config.SYSConf.WatcherInterval = WatcherInterval
	}

	// 解密 Trap Community, 配置中的环境变量名优先, 默认与 SNMP Community 相同
	envName = config.TrapConf.CommunityName
	if envName == "" {
		envName = TrapCommunityName
	}
	config.TrapConf.Community = utils.GetenvDecrypt(envName, config.SYSConf.BaseSecretValue)
	if config.TrapConf.Community == "" {
		config.TrapConf.Community = config.SNMPConf.Community
	}

	if config.TrapConf.AlarmCode == "" {
		config.TrapConf.AlarmCode = TrapAlarmCode
	}

	if config.TrapConf.LocalAddr == "" {
		config.TrapConf.LocalAddr = TrapServerAddr
	}

	if len(config.TrapConf.Identifiers) == 0 || len(config.TrapConf.InterfacePrefix) == 0 {
		return nil, fmt.Errorf("identifiers or interface prefix cannot be empty")
	}

	// 转换 Oid 类型
	config.TrapConf.InterfacePrefixOids, err = oidsSlice(config.TrapConf.InterfacePrefix)
	if err != nil {
		return nil, fmt.Errorf("interface prefix err: %v", err)
	}

	return config, nil
}

// 转换 oid 列表为 Oid 类型
func oidsSlice(oids []string) (ret snmpgo.Oids, err error) {
	for _, v := range oids {
		oid, err := snmpgo.NewOid(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		ret = append(ret, oid)
	}

	return
}

// 解析 Address Config
func parseAddrConf(conf []string, defCommuntiy string,
	v2Target map[string]TV2Target, v2TargetNew map[string]struct{}) error {
	for _, item := range conf {
		// 排除空白行, __ 或 # 开头的注释行
		item = strings.TrimSpace(item)
		if item == "" || strings.HasPrefix(item, "__") || strings.HasPrefix(item, "#") {
			continue
		}

		community := defCommuntiy
		addr := item
		port := SNMPPort

		// password123@12.3.4.5#16161
		parseItem := strings.SplitN(strings.Trim(item, "@"), "@", 2)
		if len(parseItem) == 2 {
			community = parseItem[0]
			addr = parseItem[1]
		}

		// 端口
		parseAddr := strings.SplitN(strings.Trim(addr, "#"), "#", 2)
		if len(parseAddr) == 2 {
			addr = parseAddr[0]
			port = parseAddr[1]
		}

		// 校验 IP
		ipAddr := getSNMPTarget(addr)
		if ipAddr == "" {
			return fmt.Errorf("v2 target address err: %s", item)
		}
		if strings.Contains(ipAddr, ":") {
			// IPv6
			addr = "[" + addr + "]"
		}

		// 目标地址中必须带端口
		key := utils.MD5Hex(community) + "@" + addr + ":" + port
		if _, ok := Config.SNMPConf.V2.Target[key]; !ok {
			// 新增的地址
			v2TargetNew[key] = struct{}{}
		}
		// 最终目标配置
		v2Target[key] = TV2Target{
			Community: community,
			Addr:      addr,
			Port:      port,
		}
	}

	return nil
}

// 检查是否为合法 SNMP 目标地址: 内网 10 段及公网 IP, 以及环回地址
func getSNMPTarget(addr string) string {
	ip := net.ParseIP(addr)
	if ip == nil || ip.Equal(net.IPv4bcast) || ip.IsUnspecified() || ip.IsMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return ""
	}

	// IPv4 私有地址, 排除 10 段
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			fallthrough
		case ip4[0] == 192 && ip4[1] == 168:
			return ""
		}
	}

	return ip.String()
}
