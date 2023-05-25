package top_list

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/rs/zerolog/log"

	"os"
	"sync"
	"time"
)

var Data36krDailyRecord d36DR
var WallStreetNewsDailyRecord wallStreetNewsDailyRecord
var WeiboHotDailyRecord weiboHotDailyRecord
var ZhihuHotDailyRecord zhihuHotDailyRecord

type d36DR struct {
	data map[string][]Data36krHot
	sync.RWMutex
}

func (d *d36DR) Add(k string, v []Data36krHot) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		d.data = map[string][]Data36krHot{}
	}
	d.data[k] = v
}

func (d *d36DR) GetData() map[string][]Data36krHot {
	d.RLock()
	d.RUnlock()
	data := make(map[string][]Data36krHot)
	for k, v := range d.data {
		data[k] = v
	}
	return data
}

type wallStreetNewsDailyRecord struct {
	data map[string][]WallStreetNews
	sync.RWMutex
}

func (d *wallStreetNewsDailyRecord) Add(k string, v []WallStreetNews) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		d.data = map[string][]WallStreetNews{}
	}
	d.data[k] = v
}

func (d *wallStreetNewsDailyRecord) GetData() map[string][]WallStreetNews {
	d.RLock()
	d.RUnlock()
	data := make(map[string][]WallStreetNews)
	for k, v := range d.data {
		data[k] = v
	}
	return data
}

type weiboHotDailyRecord struct {
	data map[string][]WeiboHot
	sync.RWMutex
}

func (d *weiboHotDailyRecord) Add(k string, v []WeiboHot) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		d.data = map[string][]WeiboHot{}
	}
	d.data[k] = v
}

func (d *weiboHotDailyRecord) GetData() map[string][]WeiboHot {
	d.RLock()
	d.RUnlock()
	data := make(map[string][]WeiboHot)
	for k, v := range d.data {
		data[k] = v
	}
	return data
}

type zhihuHotDailyRecord struct {
	data map[string][]ZhihuHot
	sync.RWMutex
}

func (z *zhihuHotDailyRecord) Add(k string, v []ZhihuHot) {
	z.Lock()
	defer z.Unlock()
	if z.data == nil {
		z.data = map[string][]ZhihuHot{}
	}
	z.data[k] = v
}

func (z *zhihuHotDailyRecord) GetData() map[string][]ZhihuHot {
	z.RLock()
	z.RUnlock()
	data := make(map[string][]ZhihuHot)
	for k, v := range z.data {
		data[k] = v
	}
	return data
}

// time.Now().Format("2006-01-02 15:04:05")
func UploadDailyRecord() {
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}

	//写微博热搜当日文件
	{
		weiboFilePath, err := file_util.WriteJsonFile(WeiboHotDailyRecord.GetData(), path, "weibo_hot", true)
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "write weibo hot daily record",
				"error":  err,
			}).Send()
		} else {
			cosPath := fmt.Sprintf("%s/%s", "weibo", time.Now().Format("2006"))
			cosFileName := fmt.Sprintf("%s_%s.json", "weibo_hot", time.Now().Format("0102"))
			if err = file_util.TCCosUpload(cosPath, cosFileName, weiboFilePath); err != nil {
				log.Error().Fields(map[string]interface{}{
					"action": "upload weibo hot daily record to tencent cos",
					"error":  err,
				}).Send()
			} else {
				WeiboHotDailyRecord.data = nil
				if err := file_util.ClearFile(weiboFilePath); err != nil {
					_ = os.Remove(weiboFilePath)
				}
			}
		}
	}

	//写华尔街资讯当日文件
	{
		wallStreetFilePath, err := file_util.WriteJsonFile(WallStreetNewsDailyRecord.GetData(), path, "wallstreet_news", true)
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "write weibo hot daily record",
				"error":  err,
			}).Send()
		} else {
			cosPath := fmt.Sprintf("%s/%s", "wallstreet", time.Now().Format("2006"))
			cosFileName := fmt.Sprintf("%s_%s.json", "wallstreet", time.Now().Format("0102"))
			if err = file_util.TCCosUpload(cosPath, cosFileName, wallStreetFilePath); err != nil {
				log.Error().Fields(map[string]interface{}{
					"action": "upload wall street news daily record to tencent cos",
					"error":  err,
				}).Send()
			} else {
				WallStreetNewsDailyRecord.data = nil
				if err := file_util.ClearFile(wallStreetFilePath); err != nil {
					_ = os.Remove(wallStreetFilePath)
				}
			}
		}
	}

	//写36氪日榜当日文件
	{
		d36krFilePath, err := file_util.WriteJsonFile(Data36krDailyRecord.GetData(), path, "36kr", true)
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "write weibo hot daily record",
				"error":  err,
			}).Send()
		} else {
			cosPath := fmt.Sprintf("%s/%s", "36kr", time.Now().Format("2006"))
			cosFileName := fmt.Sprintf("%s_%s.json", "36kr", time.Now().Format("0102"))
			if err = file_util.TCCosUpload(cosPath, cosFileName, d36krFilePath); err != nil {
				log.Error().Fields(map[string]interface{}{
					"action": "upload wall street news daily record to tencent cos",
					"error":  err,
				}).Send()
			} else {
				Data36krDailyRecord.data = nil
				if err := file_util.ClearFile(d36krFilePath); err != nil {
					_ = os.Remove(d36krFilePath)
				}
			}
		}
	}

	//写知乎热榜当日文件
	{
		// cause zhihu hot not register as corn job so load data by this time incase it has no data
		_, _ = LoadZhihuHot()
		zhihuFilePath, err := file_util.WriteJsonFile(ZhihuHotDailyRecord.GetData(), path, "zhihu", true)
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "write zhihu hot daily record",
				"error":  err,
			}).Send()
		} else {
			cosPath := fmt.Sprintf("%s/%s", "zhihu", time.Now().Format("2006"))
			cosFileName := fmt.Sprintf("%s_%s.json", "zhihu", time.Now().Format("0102"))
			if err = file_util.TCCosUpload(cosPath, cosFileName, zhihuFilePath); err != nil {
				log.Error().Fields(map[string]interface{}{
					"action": "upload wall street news daily record to tencent cos",
					"error":  err,
				}).Send()
			} else {
				Data36krDailyRecord.data = nil
				if err := file_util.ClearFile(zhihuFilePath); err != nil {
					_ = os.Remove(zhihuFilePath)
				}
			}
		}
	}
}

func init() {
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}

	//加载微博每日记录
	{
		data := make(map[string][]WeiboHot)
		if err := file_util.LoadJsonFile(fmt.Sprintf("%s/weibo_hot.json", path), &data); err == nil {
			WeiboHotDailyRecord = weiboHotDailyRecord{
				data:    data,
				RWMutex: sync.RWMutex{},
			}
		}

	}

	//加载华尔街每日记录
	{
		data := make(map[string][]WallStreetNews)
		if err := file_util.LoadJsonFile(fmt.Sprintf("%s/wallstreet_news.json", path), &data); err == nil {
			WallStreetNewsDailyRecord = wallStreetNewsDailyRecord{
				data:    data,
				RWMutex: sync.RWMutex{},
			}
		}
	}

	//加载36氪每日记录
	{
		data := make(map[string][]Data36krHot)
		if err := file_util.LoadJsonFile(fmt.Sprintf("%s/36kr.json", path), &data); err == nil {
			Data36krDailyRecord = d36DR{
				data:    data,
				RWMutex: sync.RWMutex{},
			}
		}
	}

	//加载知乎热榜每日记录
	{
		data := make(map[string][]ZhihuHot)
		if err := file_util.LoadJsonFile(fmt.Sprintf("%s/zhihu.json", path), &data); err == nil {
			ZhihuHotDailyRecord = zhihuHotDailyRecord{
				data:    data,
				RWMutex: sync.RWMutex{},
			}
		}
	}

}
