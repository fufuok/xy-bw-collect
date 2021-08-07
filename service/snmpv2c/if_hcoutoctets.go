package snmpv2c

import (
	"strconv"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 获取流出带宽
func getIfHCOutOctets(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IfHCOutOctets})
	if err != nil {
		return err
	}

	depth := len(common.IfHCOutOctets.Value)
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IfHCOutOctets) {
		if len(varBind.Oid.Value) < depth+1 {
			continue
		}

		ifNum := varBind.Oid.Value[depth]
		oid, err := common.IfHCOutOctets.AppendSubIds([]int{ifNum})
		if err != nil {
			return err
		}

		var count uint64
		if data := pdu.VarBinds().MatchOid(oid); data != nil {
			count, err = strconv.ParseUint(data.Variable.String(), 10, 64)
			if err != nil {
				return err
			}
		}
		info.LastIfOut[ifNum], info.IfOut[ifNum] = info.IfOut[ifNum], count
	}

	// 更新上一次日期
	info.LastTime["ifout"], info.Time["ifout"] = info.Time["ifout"], time.Now()

	return nil
}
