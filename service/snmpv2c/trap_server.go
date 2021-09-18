package snmpv2c

import (
	"log"
	"strings"
	"time"

	"github.com/fufuok/cache2go"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xid"
	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
)

type tTrapListener struct{}

func (l *tTrapListener) OnTRAP(trap *snmpgo.TrapRequest) {
	if trap.Error != nil {
		common.Log.Error().Err(trap.Error).Str("src_addr", trap.Source.String()).Msg("OnTRAP")
		return
	}

	var ifName, event string

	for _, val := range trap.Pdu.VarBinds() {
		switch val.Variable.Type() {
		case "Oid":
			// 匹配标识符描述名称
			if oidDescr, ok := conf.Config.TrapConf.Identifiers[val.Variable.String()]; ok {
				event = oidDescr
				continue
			}
		case "OctetString":
			// 匹配限定收集的 Oids 前缀
			for _, oid := range conf.Config.TrapConf.InterfacePrefixOids {
				if val.Oid.Contains(oid) {
					ifName += "," + val.Variable.String()
				}
			}
		}
	}

	if ifName != "" && event != "" {
		table := time.Now().Format(conf.TrapCacheTable)
		cache := cache2go.Cache(table)

		ifName = strings.Trim(ifName, ",")
		srcIP, _, _ := utils.GetIPPort(trap.Source)
		nodeIP := srcIP.String()
		nodeIPName := utils.GetString(conf.Config.TrapConf.IPNameTable[nodeIP], nodeIP)

		now := time.Now()
		data := utils.MustJSON(map[string]interface{}{
			"code":      conf.Config.TrapConf.AlarmCode,
			"node_ip":   nodeIP,
			"ifname":    ifName,
			"event":     event,
			"timestamp": now.Unix(),
			"time":      now.Format(time.RFC3339),
			"info":      nodeIPName,
		})

		if conf.Debug {
			common.LogSampled.Debug().Bytes("data", data).Msg("TRAP")
		}

		key := xid.NewString()
		cache.Add(key, 5*time.Second, data)
	}
}

// InitTrapServer 初始化 Trap 监听服务
func InitTrapServer() {
	common.Log.Info().
		Str("addr", conf.Config.TrapConf.LocalAddr).
		Msg("Listening and serving TRAP")
	if err := newTrapServer(); err != nil {
		log.Fatalln("Failed to start Trap Server:", err, "\nbye.")
	}
}

// 启动 TrapServer
func newTrapServer() error {
	svr, err := snmpgo.NewTrapServer(snmpgo.ServerArguments{
		LocalAddr: conf.Config.TrapConf.LocalAddr,
	})
	if err != nil {
		return err
	}

	// v2c
	err = svr.AddSecurity(&snmpgo.SecurityEntry{
		Version:   snmpgo.V2c,
		Community: conf.Config.TrapConf.Community,
	})
	if err != nil {
		return err
	}

	err = svr.Serve(&tTrapListener{})
	if err != nil {
		return err
	}

	return nil
}
