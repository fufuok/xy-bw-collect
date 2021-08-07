package snmpv2c

import (
	"fmt"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
)

// 获取网卡名称
func getIfName(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IfName})
	if err != nil {
		return err
	}

	banRegexp := conf.Config.SNMPConf.BanRegexp
	depth := len(common.IfName.Value)
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IfName) {
		if len(varBind.Oid.Value) < depth+1 {
			continue
		}

		ifNum := varBind.Oid.Value[depth]
		ifName := varBind.Variable.String()

		if banRegexp != nil && banRegexp.MatchString(ifName) {
			// 排除的接口
			continue
		}

		// 更新网卡信息
		info.IfName[ifNum] = ifName
	}

	if len(info.IfName) == 0 {
		return fmt.Errorf("no interface: %s", snmpOption.Address)
	}

	// 更新上一次日期
	info.LastTime["ifname"], info.Time["ifname"] = info.Time["ifname"], time.Now()

	return nil
}
