package common

import (
	"time"

	"github.com/fufuok/utils/xid"
	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/conf"
)

type TIFInfo struct {
	GoID          string
	IfAdminStatus map[int]string
	IfName        map[int]string
	IpAddr        map[int]string
	IfIn          map[int]uint64
	IfOut         map[int]uint64
	LastIfIn      map[int]uint64
	LastIfOut     map[int]uint64
	IfInPkts      map[int]uint64
	IfOutPkts     map[int]uint64
	LastIfInPkts  map[int]uint64
	LastIfOutPkts map[int]uint64
	Time          map[string]time.Time
	LastTime      map[string]time.Time
	Duration      time.Duration
}

type TBWReport struct {
	Ip          string  `json:"ip"`
	In          uint64  `json:"kbps_in"`
	Out         uint64  `json:"kbps_out"`
	PPSIn       uint64  `json:"pps_in"`
	PPSOut      uint64  `json:"pps_out"`
	Time        string  `json:"time"`
	TargetIP    string  `json:"client_ip"`
	Interface   string  `json:"interface"`
	IfNum       int     `json:"interface_num"`
	RawIn       string  `json:"raw_in"`
	RawOut      string  `json:"raw_out"`
	Timestamp   int64   `json:"timestamp"`
	Duration    string  `json:"duration"`
	Interval    float64 `json:"interval"`
	HumanIn     string  `json:"human_bps_in"`
	HumanOut    string  `json:"human_bps_out"`
	CommaPPSIn  string  `json:"comma_pps_in"`
	CommaPPSOut string  `json:"comma_pps_out"`

	conf.TESIPs
}

// var SYSDescr = snmpgo.MustNewOid("1.3.6.1.2.1.1.1")

var IfAdminStatus = snmpgo.MustNewOid("1.3.6.1.2.1.2.2.1.7")
var IfName = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.1")
var IfHCInOctets = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.6")
var IfHCOutOctets = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.10")

var IfHCInUcastPkts = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.7")

// var IfHCInMulticastPkts = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.8")
// var IfHCInBroadcastPkts = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.9")

var IfHCOutUcastPkts = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.11")

// var IfHCOutMulticastPkts = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.12")
// var IfHCOutBroadcastPkts = snmpgo.MustNewOid("1.3.6.1.2.1.31.1.1.1.13")

// var IfMtu = snmpgo.MustNewOid("1.3.6.1.2.1.2.2.1.4")
// var IfPhysAddress = snmpgo.MustNewOid("1.3.6.1.2.1.2.2.1.6")

var IpAdEntIfIndex = snmpgo.MustNewOid("1.3.6.1.2.1.4.20.1.2")

func NewIFInfo() *TIFInfo {
	// Time["any"] is "0001-01-01 00:00:00 +0000 UTC"
	return &TIFInfo{
		GoID:          xid.NewString() + time.Now().Format(".060102150405"),
		IfAdminStatus: make(map[int]string),
		IfName:        make(map[int]string),
		IpAddr:        make(map[int]string),
		IfIn:          make(map[int]uint64),
		IfOut:         make(map[int]uint64),
		LastIfIn:      make(map[int]uint64),
		LastIfOut:     make(map[int]uint64),
		IfInPkts:      make(map[int]uint64),
		IfOutPkts:     make(map[int]uint64),
		LastIfInPkts:  make(map[int]uint64),
		LastIfOutPkts: make(map[int]uint64),
		Time:          make(map[string]time.Time),
		LastTime:      make(map[string]time.Time),
	}
}

func InitCommon() {
	// 初始化日志环境
	initLogger()
}
