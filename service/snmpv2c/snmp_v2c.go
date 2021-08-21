package snmpv2c

import (
	"fmt"
	"time"

	"github.com/fufuok/cache2go"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xid"
	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
	"github.com/fufuok/xy-bw-collect/internal/json"
)

// Collector SNMP 采集器 v2c
func Collector(target conf.TV2Target, info *common.TIFInfo) {
	snmpOption := snmpgo.SNMPArguments{
		Version:   snmpgo.V2c,
		Address:   target.Addr + ":" + target.Port,
		Retries:   conf.Config.SNMPConf.Retries,
		Community: target.Community,
		Timeout:   time.Duration(conf.Config.SNMPConf.Timeout) * time.Second,
	}

	// 收集数据
	if msg, err := doCollect(snmpOption, info); err != nil {
		// 附带 ES 源 IP 数据
		esIP, _ := conf.ESIPs.GetValue(target.Addr).(conf.TESIPs)
		common.Log.Error().Err(err).
			Str("addr", snmpOption.Address).
			Uint("retries", snmpOption.Retries).
			Dur("timeout", snmpOption.Timeout).
			Str("node_ip", esIP.NodeIP).
			Str("snmp_ip", esIP.SNMPIP).
			Msg(msg)
		return
	}

	// 采集调试信息
	if conf.Config.SYSConf.Debug {
		tmpInfo, _ := json.Marshal(info)
		common.Log.Debug().Msgf("\n\ninfo[%s - %s:%s]: \n%s\n%s\n",
			info.GoID, target.Addr, target.Port, string(tmpInfo), info.Duration)
	}

	if info.LastTime["ifin"].Year() == 1 || len(info.IfName) == 0 || len(info.IpAddr) == 0 {
		// 信息采集有误, 或首次采集数据
		return
	}

	// 整合数据, 提交数据
	dataMerge(target.Addr, info)
}

// 收集数据
func doCollect(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) (string, error) {
	if info.Time["ifin"].Minute()/30 == 0 {
		// 首次或每 30 分钟更新接口信息
		if err := getIfInfo(snmpOption, info); err != nil {
			return "getIfInfo", err
		}
	}

	// 采集流量
	if err := getIfTraffic(snmpOption, info); err != nil {
		return "getIfTraffic", err
	}

	return "", nil
}

// 合并数据, 提交到队列
func dataMerge(addr string, info *common.TIFInfo) {
	// 附带 ES 源 IP 数据
	esIP, _ := conf.ESIPs.GetValue(addr).(conf.TESIPs)

	// 丢弃超时数据
	interval := info.Time["ifin"].Sub(info.LastTime["ifin"]).Seconds()
	if info.Duration > conf.Config.SNMPConf.DiscardDuration ||
		interval < conf.Config.SNMPConf.DiscardIntervalMin ||
		interval > conf.Config.SNMPConf.DiscardIntervalMax {
		common.Log.Error().
			Str("addr", addr).
			Str("node_ip", esIP.NodeIP).
			Str("snmp_ip", esIP.SNMPIP).
			Float64("interval", interval).
			Dur("duration", info.Duration).
			Msg("discard value")
		return
	}

	table := time.Now().Format(conf.BWCacheTable)
	cache := cache2go.Cache(table)

	for ifNum := range info.IfName {
		// 流入 kbps/s
		if info.IfIn[ifNum] < info.LastIfIn[ifNum] {
			// 意外情况, 网卡重置或数据异常
			continue
		}
		kbpsIn := uint64(float64(info.IfIn[ifNum]-info.LastIfIn[ifNum]) * 8 / 1000 / interval)

		// 流出 kbps/s
		if info.IfOut[ifNum] < info.LastIfOut[ifNum] {
			continue
		}
		kbpsOut := uint64(float64(info.IfOut[ifNum]-info.LastIfOut[ifNum]) * 8 / 1000 / interval)

		// 流入 PPS
		if info.IfInPkts[ifNum] < info.LastIfInPkts[ifNum] {
			continue
		}
		ppsIn := uint64(float64(info.IfInPkts[ifNum]-info.LastIfInPkts[ifNum]) / interval)

		// 流出 PPS
		if info.IfOutPkts[ifNum] < info.LastIfOutPkts[ifNum] {
			continue
		}
		ppsOut := uint64(float64(info.IfOutPkts[ifNum]-info.LastIfOutPkts[ifNum]) / interval)

		// 最终结果数据
		bwData := utils.MustJSON(common.TBWReport{
			Ip:          info.IpAddr[ifNum],
			In:          kbpsIn,
			Out:         kbpsOut,
			PPSIn:       ppsIn,
			PPSOut:      ppsOut,
			Time:        info.Time["ifin"].Format(time.RFC3339),
			TargetIP:    addr,
			Interface:   info.IfName[ifNum],
			IfNum:       ifNum,
			RawIn:       utils.MustString(info.IfIn[ifNum]),
			RawOut:      utils.MustString(info.IfOut[ifNum]),
			Timestamp:   info.Time["ifin"].Unix(),
			Duration:    info.Duration.String(),
			Interval:    interval,
			TESIPs:      esIP,
			HumanIn:     utils.HumanKbps(kbpsIn * 1000),
			HumanOut:    utils.HumanKbps(kbpsOut * 1000),
			CommaPPSIn:  utils.Commau(ppsIn),
			CommaPPSOut: utils.Commau(ppsOut),
		})

		// 采集调试信息
		if conf.Config.SYSConf.Debug {
			common.Log.Debug().Msgf("\n\nbwData[%s]: \n%s\n", info.GoID, string(bwData))
		}

		key := xid.NewString()
		cache.Add(key, 3*time.Minute, bwData)
	}
}

// 获取数据
func getBulkWalk(snmpOption snmpgo.SNMPArguments, oids []*snmpgo.Oid) (snmpgo.Pdu, error) {
	snmp, err := snmpgo.NewSNMP(snmpOption)
	if err != nil {
		return nil, fmt.Errorf("snmp create, %s", err)
	}
	pdu, err := snmp.GetBulkWalk(oids, 0, conf.Config.SNMPConf.MaxRepetitions)
	if err != nil {
		return nil, fmt.Errorf("getbulkwalk, %s", err)
	}
	if pdu.ErrorStatus() != snmpgo.NoError {
		return nil, fmt.Errorf("snmp get data, %s(%d)", pdu.ErrorStatus(), pdu.ErrorIndex())
	}

	return pdu, nil
}
