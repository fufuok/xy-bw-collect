package conf

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
)

type tMonitorSource struct {
	OK   int           `json:"ok"`
	Msg  string        `json:"msg"`
	Data []interface{} `json:"data"`
}

type tESSource struct {
	OK   int                    `json:"ok"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

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

	var res tMonitorSource
	if err := resp.ToJSON(&res); err != nil {
		return err
	}

	if res.OK != 1 {
		return fmt.Errorf("数据源获取失败: %s", res.Msg)
	}

	// 获取所有配置项数据
	body := ""
	for _, x := range res.Data {
		info, ok := x.(map[string]interface{})
		if ok {
			if txt, ok := info["ip_info"]; ok {
				body += utils.MustString(txt) + "\n"
			}
		}
	}

	if body == "" {
		return fmt.Errorf("数据源获取结果为空")
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

	var res tESSource
	if err := resp.ToJSON(&res); err != nil {
		return err
	}

	if res.OK != 1 {
		return fmt.Errorf("数据源获取失败: %s", res.Msg)
	}

	hits, ok := res.Data["hits"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("数据源获取结果为空")
	}

	body := ""
	hitsData, _ := hits["hits"].([]interface{})
	for _, v := range hitsData {
		d, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		source, ok := d["_source"].(map[string]interface{})
		if !ok {
			continue
		}
		esIP := TESIPs{
			NodeIP: getSNMPTarget(utils.MustString(source["node_ip"])),
			SNMPIP: getSNMPTarget(utils.MustString(source["snmp_ip"])),
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

	if body == "" {
		return fmt.Errorf("数据源获取结果为空")
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
