package snmpv2c

import (
	"strconv"
	"strings"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 获取 IP
func getIpAdEntIfIndex(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{common.IpAdEntIfIndex})
	if err != nil {
		return err
	}

	// IP, 接口多 IP 时使用最后一个 IP
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IpAdEntIfIndex) {
		if varBind.Variable.Type() != "Integer" {
			continue
		}
		ifNum, _ := strconv.Atoi(varBind.Variable.String())
		info.IpAddr[ifNum] = strings.TrimPrefix(varBind.Oid.String(), common.IpAdEntIfIndex.String()+".")
	}

	return nil
}
