package conf

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

// GetMonitorSource 获取监控平台源数据配置
func (c *TFilesConf) GetMonitorSource() error {
	// Token: md5(timestamp + auth_key)
	timestamp := utils.MustString(time.Now().Unix())
	token := utils.MD5Hex(timestamp + c.SecretValue)

	// 请求数据源
	resp, err := req.Get(c.API+token+"&time="+timestamp, ReqUserAgent)
	if err != nil {
		return err
	}

	res := resp.String()
	if gjson.Get(res, "ok").Bool() {
		// 获取所有配置项数据 ["item1 ip txt", "item2 ip txt"]
		body := ""
		data := gjson.Get(res, "data.#.ip_info").Array()
		if len(data) > 0 {
			for _, x := range data {
				body += x.String() + "\n"
			}

			md5New := utils.MD5Hex(body)
			if md5New != c.ConfigMD5 {
				// 保存到配置文件
				if err := ioutil.WriteFile(c.Path, []byte(body), 0644); err != nil {
					return err
				}
				c.ConfigMD5 = md5New
				c.ConfigVer = time.Now()
			}

			return nil
		}
	}

	return fmt.Errorf("数据源获取失败: %s", gjson.Get(res, "msg").String())
}

// GetESSource 获取 ES 数据, snmpip
func (c *TFilesConf) GetESSource() error {
	// ES 查询参数
	params := req.Param{
		"index": c.ESIndex + "_" + time.Now().Format("060102"),
		"body":  c.ESBody,
	}

	// 请求 ES 数据
	resp, err := req.Post(c.API, req.BodyJSON(params), ReqUserAgent)
	if err != nil {
		return err
	}

	res := resp.String()
	if gjson.Get(res, "ok").Bool() {
		body := ""
		data := gjson.Get(res, "data.hits.hits.#._source").Array()
		if len(data) > 0 {
			for _, x := range data {
				esIP := TESIPs{
					NodeIP: getSNMPTarget(x.Get("node_ip").String()),
					SNMPIP: getSNMPTarget(x.Get("snmp_ip").String()),
				}

				// 优先环网 IP 10.0.0.0/8
				ip := esIP.SNMPIP
				if ip == "" {
					ip = esIP.NodeIP
					if ip == "" {
						continue
					}
				}

				// 文件配置
				body += ip + "\n"

				// 采集目标 IP 对应的 ES 源 IP 数据
				ESIPs.Set(ip, esIP)
			}

			md5New := utils.MD5Hex(body)
			if md5New != c.ConfigMD5 {
				// 保存到配置文件
				if err := ioutil.WriteFile(c.Path, []byte(body), 0644); err != nil {
					return err
				}
				c.ConfigMD5 = md5New
				c.ConfigVer = time.Now()
			}

			return nil
		}
	}

	return fmt.Errorf("数据源获取失败: %s", gjson.Get(res, "msg").String())
}
