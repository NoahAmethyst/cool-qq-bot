package top_list

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/rs/zerolog/log"

	"os"
	"sync"
	"time"
)

var D36krDailyRecord d36KrDailyRecord
var WallStreetNewsDailyRecord wallStreetNewsDailyRecord
var WeiboHotDailyRecord weiboHotDailyRecord
var ZhihuHotDailyRecord zhihuHotDailyRecord

type DailyRecord interface {
	Upload()
	Load()
}

type d36KrDailyRecord struct {
	data map[string][]Data36krHot
	sync.RWMutex
}

func (d *d36KrDailyRecord) Add(k string, v []Data36krHot) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		d.data = map[string][]Data36krHot{}
	}
	d.data[k] = v
}

func (d *d36KrDailyRecord) GetData() map[string][]Data36krHot {
	d.RLock()
	d.RUnlock()
	data := make(map[string][]Data36krHot)
	for k, v := range d.data {
		data[k] = v
	}
	return data
}

// 写36氪日榜当日文件
func (d *d36KrDailyRecord) Upload() {
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()

	d36krFilePath, err := file_util.WriteJsonFile(d.data, path, "36kr", true)
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
			d.data = nil
			if err := file_util.ClearFile(d36krFilePath); err != nil {
				_ = os.Remove(d36krFilePath)
			}
		}
	}

}

// 加载36氪每日记录
func (d *d36KrDailyRecord) Load() {
	d.RWMutex = sync.RWMutex{}
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()

	data := make(map[string][]Data36krHot)
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/36kr.json", path), &data); err == nil {
		d.data = data

	}

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

// 写华尔街资讯当日文件
func (d *wallStreetNewsDailyRecord) Upload() {
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()
	wallStreetFilePath, err := file_util.WriteJsonFile(d.data, path, "wallstreet_news", true)
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
			d.data = nil
			if err := file_util.ClearFile(wallStreetFilePath); err != nil {
				_ = os.Remove(wallStreetFilePath)
			}
		}
	}
}

// 加载华尔街每日记录
func (d *wallStreetNewsDailyRecord) Load() {
	d.RWMutex = sync.RWMutex{}
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()
	data := make(map[string][]WallStreetNews)
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/wallstreet_news.json", path), &data); err == nil {
		d.data = data
	}
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

// 写微博热搜当日文件
func (d *weiboHotDailyRecord) Upload() {
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()
	weiboFilePath, err := file_util.WriteJsonFile(d.data, path, "weibo_hot", true)
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
			d.data = nil
			if err := file_util.ClearFile(weiboFilePath); err != nil {
				_ = os.Remove(weiboFilePath)
			}
		}
	}
}

// 加载微博每日记录
func (d *weiboHotDailyRecord) Load() {
	d.RWMutex = sync.RWMutex{}
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()
	data := make(map[string][]WeiboHot)
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/weibo_hot.json", path), &data); err == nil {
		d.data = data

	}

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

// 写知乎热榜当日文件
func (z *zhihuHotDailyRecord) Upload() {
	z.Lock()
	defer z.Unlock()
	path := file_util.GetFileRoot()

	// cause zhihu hot not register as corn job so load data by this time incase it has no data
	_, _ = LoadZhihuHot()
	zhihuFilePath, err := file_util.WriteJsonFile(z.data, path, "zhihu", true)
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
			z.data = nil
			if err := file_util.ClearFile(zhihuFilePath); err != nil {
				_ = os.Remove(zhihuFilePath)
			}
		}
	}

}

// 加载知乎热榜每日记录
func (z *zhihuHotDailyRecord) Load() {
	z.RWMutex = sync.RWMutex{}
	z.Lock()
	defer z.Unlock()
	path := file_util.GetFileRoot()
	data := make(map[string][]ZhihuHot)
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/zhihu.json", path), &data); err == nil {
		z.data = data
	}

}

// time.Now().Format("2006-01-02 15:04:05")
func UploadDailyRecord() {

	//写微博热搜当日文件
	WeiboHotDailyRecord.Upload()

	//写华尔街资讯当日文件
	WallStreetNewsDailyRecord.Upload()

	//写36氪日榜当日文件
	D36krDailyRecord.Upload()

	//写知乎热榜当日文件
	ZhihuHotDailyRecord.Upload()
}

func init() {
	//加载微博热搜每日记录
	WeiboHotDailyRecord.Load()

	//加载华尔街每日记录
	WallStreetNewsDailyRecord.Load()

	//加载36氪每日记录
	D36krDailyRecord.Load()

	//加载知乎热榜每日记录
	ZhihuHotDailyRecord.Load()

}
