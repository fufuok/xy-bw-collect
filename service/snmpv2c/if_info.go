package snmpv2c

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/fufuok/xy-bw-collect/common"
	"github.com/fufuok/xy-bw-collect/conf"
)

// 获取接口信息
func getIfInfo(snmpOption snmpgo.SNMPArguments, info *common.TIFInfo) error {
	pdu, err := getBulkWalk(snmpOption, snmpgo.Oids{
		common.IfAdminStatus,
		common.IfName,
		common.IpAdEntIfIndex,
	})
	if err != nil {
		return err
	}

	newIfName := make(map[int]string)
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

		oid, err := common.IfAdminStatus.AppendSubIds([]int{ifNum})
		if err != nil {
			return err
		}
		if adms := pdu.VarBinds().MatchOid(oid); adms == nil || adms.Variable.String() != "1" {
			// 网卡状态 1 为启用
			continue
		}

		// 更新网卡信息
		newIfName[ifNum] = ifName
	}

	if len(newIfName) == 0 {
		return fmt.Errorf("no interface, %s", snmpOption.Address)
	}

	info.IfName = newIfName

	// IP, 接口多 IP 时使用最后一个 IP
	for _, varBind := range pdu.VarBinds().MatchBaseOids(common.IpAdEntIfIndex) {
		if varBind.Variable.Type() != "Integer" {
			continue
		}
		ifNum, _ := strconv.Atoi(varBind.Variable.String())
		info.IpAddr[ifNum] = strings.TrimPrefix(varBind.Oid.String(), common.IpAdEntIfIndex.String()+".")
	}

	// 更新上一次日期
	info.LastTime["ifinfo"], info.Time["ifinfo"] = info.Time["ifinfo"], time.Now()

	return nil
}
