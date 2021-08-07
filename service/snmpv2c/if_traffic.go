package snmpv2c

import (
	"strconv"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
)

// 采集流量
func getIfTraffic(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	start := time.Now()
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{
		common.IfHCInOctets,
		common.IfHCOutOctets,
		common.IfHCInUcastPkts,
		common.IfHCOutUcastPkts,
	})
	if err != nil {
		return err
	}
	end := time.Now()

	// 更新上一次日期
	info.LastTime["ifin"], info.Time["ifin"] = info.Time["ifin"], end
	info.Duration = end.Sub(start)

	for ifNum := range info.IfName {
		// 流入带宽
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

		// 流出带宽
		oid, err = common.IfHCOutOctets.AppendSubIds([]int{ifNum})
		if err != nil {
			return err
		}

		count = 0
		if data := pdu.VarBinds().MatchOid(oid); data != nil {
			count, err = strconv.ParseUint(data.Variable.String(), 10, 64)
			if err != nil {
				return err
			}
		}
		info.LastIfOut[ifNum], info.IfOut[ifNum] = info.IfOut[ifNum], count

		// 流入 PPS
		oid, err = common.IfHCInUcastPkts.AppendSubIds([]int{ifNum})
		if err != nil {
			return err
		}

		count = 0
		if data := pdu.VarBinds().MatchOid(oid); data != nil {
			count, err = strconv.ParseUint(data.Variable.String(), 10, 64)
			if err != nil {
				return err
			}
		}
		info.LastIfInPkts[ifNum], info.IfInPkts[ifNum] = info.IfInPkts[ifNum], count

		// 流出 PPS
		oid, err = common.IfHCOutUcastPkts.AppendSubIds([]int{ifNum})
		if err != nil {
			return err
		}

		count = 0
		if data := pdu.VarBinds().MatchOid(oid); data != nil {
			count, err = strconv.ParseUint(data.Variable.String(), 10, 64)
			if err != nil {
				return err
			}
		}
		info.LastIfOutPkts[ifNum], info.IfOutPkts[ifNum] = info.IfOutPkts[ifNum], count
	}

	return nil
}
