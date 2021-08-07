package snmpv2c

import (
	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 获取网卡状态
func getIfAdminStatus(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IfAdminStatus})
	if err != nil {
		return err
	}

	depth := len(common.IfAdminStatus.Value)
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IfAdminStatus) {
		if len(varBind.Oid.Value) < depth+1 {
			continue
		}
		ifNum := varBind.Oid.Value[depth]
		info.IfAdminStatus[ifNum] = varBind.Variable.String()
	}

	return nil
}
