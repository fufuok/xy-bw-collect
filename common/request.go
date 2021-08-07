package common

import (
	"bytes"
	"time"

	"github.com/fufuok/cache2go"
	"github.com/fufuok/utils"
	"github.com/imroc/req"

	"github.com/fufuok/xy-bw-collect/conf"
)

var bodySep = []byte(conf.ESBodySep)

// SendLastData 发送相应时间范围的数据到 ES
func SendLastData(dur time.Duration, key string, apiUrl string) {
	var bodyBuf bytes.Buffer
	i := 0
	cacheName := time.Now().Add(dur).Format(key)
	cache := cache2go.Cache(cacheName)
	cache.Foreach(func(_ interface{}, item *cache2go.CacheItem) {
		bodyBuf.Write(utils.GetBytes(item.Data()))
		bodyBuf.Write(bodySep)
		i = i + 1
		// 按内容大小或条数分送发送
		if i%conf.ESPostBatchNum == 0 || bodyBuf.Len() > conf.ESPOSTBatchBytes {
			_, _ = req.Post(apiUrl, req.BodyJSON(&bodyBuf), conf.ReqUserAgent)
			bodyBuf.Reset()
			i = 0
		}
	})
	if i > 0 {
		_, _ = req.Post(apiUrl, req.BodyJSON(&bodyBuf), conf.ReqUserAgent)
	}

	// 使用完毕销毁缓存表
	cache.Release()
}
