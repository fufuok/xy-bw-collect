# XY.BWCollect (SNMP 带宽采集和 Trap Server)

## 功能

- 通过 SNMP 采集网络设备带宽并上报到 ES, 每分钟一次
- 监听 SNMP Trap, 发送消息到报警平台

## 安全

- 内网 IP 采集
- IP 白名单
- SNMPv2c Community

## 日志

见: [CHANGELOG.md](CHANGELOG.md)

## 说明

1. 运行 `./xybwcollect` 默认使用配置文件目录下 `../etc/xybwcollect.json`
2. 可以指定配置文件运行 `./xybwcollect -c /mydir/conf.json`
3. 自动后台运行并守护自身, 守护日志在 `log/daemon.log`, `Error` 日志存放于 `log` 并发到日志接口





*ff*