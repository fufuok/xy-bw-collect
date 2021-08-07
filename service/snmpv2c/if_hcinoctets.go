package snmpv2c

import (
	"strconv"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 获取流入带宽
func getIfHCInOctets(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IfHCInOctets})
	if err != nil {
		return err
	}

	depth := len(common.IfHCInOctets.Value)
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IfHCInOctets) {
		if len(varBind.Oid.Value) < depth+1 {
			continue
		}

		ifNum := varBind.Oid.Value[depth]
		oid, err := common.IfHCInOctets.AppendSubIds([]int{ifNum})
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
		info.LastIfIn[ifNum], info.IfIn[ifNum] = info.IfIn[ifNum], count
	}

	// 更新上一次日期
	info.LastTime["ifin"], info.Time["ifin"] = info.Time["ifin"], time.Now()

	return nil
}
