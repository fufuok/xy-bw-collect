package snmpv2c

import (
	"strconv"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 获取流出 PPS
func getIfHCOutUcastPkts(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IfHCOutUcastPkts})
	if err != nil {
		return err
	}

	depth := len(common.IfHCOutUcastPkts.Value)
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IfHCOutUcastPkts) {
		if len(varBind.Oid.Value) < depth+1 {
			continue
		}

		ifNum := varBind.Oid.Value[depth]
		oid, err := common.IfHCOutUcastPkts.AppendSubIds([]int{ifNum})
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
		info.LastIfOutPkts[ifNum], info.IfOutPkts[ifNum] = info.IfOutPkts[ifNum], count
	}

	// 更新上一次日期
	info.LastTime["ifoutpkts"], info.Time["ifoutpkts"] = info.Time["ifoutpkts"], time.Now()

	return nil
}
