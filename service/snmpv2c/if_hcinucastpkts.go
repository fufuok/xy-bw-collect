package snmpv2c

import (
	"strconv"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 获取流入 PPS
func getIfHCInUcastPkts(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IfHCInUcastPkts})
	if err != nil {
		return err
	}

	depth := len(common.IfHCInUcastPkts.Value)
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IfHCInUcastPkts) {
		if len(varBind.Oid.Value) < depth+1 {
			continue
		}

		ifNum := varBind.Oid.Value[depth]
		oid, err := common.IfHCInUcastPkts.AppendSubIds([]int{ifNum})
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
		info.LastIfInPkts[ifNum], info.IfInPkts[ifNum] = info.IfInPkts[ifNum], count
	}

	// 更新上一次日期
	info.LastTime["ifinpkts"], info.Time["ifinpkts"] = info.Time["ifinpkts"], time.Now()

	return nil
}
