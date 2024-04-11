package top_list

import (
	"fmt"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	"github.com/NoahAmethyst/go-cqhttp/util/encrypt"
	"github.com/NoahAmethyst/go-cqhttp/util/file_util"
	log "github.com/sirupsen/logrus"

	"os"
	"sync"
	"time"
)

var D36krDailyRecord d36KrDailyRecord
var WallStreetNewsDailyRecord wallStreetNewsDailyRecord
var CaiXinNewsDailyRecord caixinnewsDailyRecord
var WeiboHotDailyRecord weiboHotDailyRecord
var ZhihuHotDailyRecord zhihuHotDailyRecord
var SentRecord sentCache

type DailyRecord interface {
	Upload()
	Load()
	Backup()
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
		log.Errorf("Write weibo hot daily record failed %s", err.Error())
	} else {
		cosPath := fmt.Sprintf("%s/%s", "36kr", time.Now().Format("2006"))
		cosFileName := fmt.Sprintf("%s_%s.json", "36kr", time.Now().Format("0102"))
		if err = file_util.TCCosUpload(cosPath, cosFileName, d36krFilePath); err != nil {
			log.Errorf("Upload wall street news daily record to tencent cos failed %s", err.Error())
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
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/36kr.json", path), &data); err != nil {
		log.Errorf("load 36kr daily record failed:%s", err.Error())
		data = make(map[string][]Data36krHot)
	}
	d.data = data
}

func (d *d36KrDailyRecord) Backup() {
	d.RLock()
	defer d.RUnlock()
	//log.Infof("backup 36kr daily record")
	path := file_util.GetFileRoot()
	_, _ = file_util.WriteJsonFile(d.data, path, "36kr", true)

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
		log.Errorf("Write wallstreet news daily record failed %s", err.Error())
	} else {
		cosPath := fmt.Sprintf("%s/%s", "wallstreet", time.Now().Format("2006"))
		cosFileName := fmt.Sprintf("%s_%s.json", "wallstreet", time.Now().Format("0102"))
		if err = file_util.TCCosUpload(cosPath, cosFileName, wallStreetFilePath); err != nil {
			log.Errorf("Upload wall street news daily record to tencent cos failed %s", err.Error())
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
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/wallstreet_news.json", path), &data); err != nil {
		log.Errorf("load wallstreet news failed:%s", err.Error())
		data = make(map[string][]WallStreetNews)
	}
	d.data = data
}

func (d *wallStreetNewsDailyRecord) Backup() {
	d.RLock()
	defer d.RUnlock()
	//log.Infof("backup wallstreet news daily record")
	path := file_util.GetFileRoot()
	_, _ = file_util.WriteJsonFile(d.data, path, "wallstreet_news", true)
}

type caixinnewsDailyRecord struct {
	data map[string][]spider_pb.CaiXinNew
	sync.RWMutex
}

func (d *caixinnewsDailyRecord) Add(k string, v []spider_pb.CaiXinNew) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		d.data = map[string][]spider_pb.CaiXinNew{}
	}
	d.data[k] = v
}

func (d *caixinnewsDailyRecord) GetData() map[string][]spider_pb.CaiXinNew {
	d.RLock()
	d.RUnlock()
	data := make(map[string][]spider_pb.CaiXinNew)
	for k, v := range d.data {
		data[k] = v
	}
	return data
}

// 写财新新闻当日文件
func (d *caixinnewsDailyRecord) Upload() {
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()
	caixinNewsFilePath, err := file_util.WriteJsonFile(d.data, path, "caixin_news", true)
	if err != nil {
		log.Errorf("Write caixin news daily record failed %s", err.Error())
	} else {
		cosPath := fmt.Sprintf("%s/%s", "caixin", time.Now().Format("2006"))
		cosFileName := fmt.Sprintf("%s_%s.json", "caixin", time.Now().Format("0102"))
		if err = file_util.TCCosUpload(cosPath, cosFileName, caixinNewsFilePath); err != nil {
			log.Errorf("Upload wall street news daily record to tencent cos failed %s", err.Error())
		} else {
			d.data = nil
			if err := file_util.ClearFile(caixinNewsFilePath); err != nil {
				_ = os.Remove(caixinNewsFilePath)
			}
		}
	}
}

// 加载财新新闻每日记录
func (d *caixinnewsDailyRecord) Load() {
	d.RWMutex = sync.RWMutex{}
	d.Lock()
	defer d.Unlock()
	path := file_util.GetFileRoot()
	data := make(map[string][]spider_pb.CaiXinNew)
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/caixin_news.json", path), &data); err != nil {
		log.Errorf("load caxin news faield:%s", err.Error())
		data = make(map[string][]spider_pb.CaiXinNew)
	}
	d.data = data
}

func (d *caixinnewsDailyRecord) Backup() {
	d.RLock()
	defer d.RUnlock()
	//log.Infof("backup caixin news daily record")
	path := file_util.GetFileRoot()
	_, _ = file_util.WriteJsonFile(d.data, path, "caixin_news", true)
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
		log.Errorf("Write weibo hot daily record failed %s", err.Error())
	} else {
		cosPath := fmt.Sprintf("%s/%s", "weibo", time.Now().Format("2006"))
		cosFileName := fmt.Sprintf("%s_%s.json", "weibo_hot", time.Now().Format("0102"))
		if err = file_util.TCCosUpload(cosPath, cosFileName, weiboFilePath); err != nil {
			log.Errorf("Upload weibo hot daily record to tencent cos failed %s", err.Error())
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
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/weibo_hot.json", path), &data); err != nil {
		log.Errorf("load weibo hot daily record failed:%s", err.Error())
		data = make(map[string][]WeiboHot)
	}
	d.data = data
}

func (d *weiboHotDailyRecord) Backup() {
	d.RLock()
	defer d.RUnlock()
	//log.Infof("backup weibo hot daily record")
	path := file_util.GetFileRoot()
	_, _ = file_util.WriteJsonFile(d.data, path, "weibo_hot", true)
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
	path := file_util.GetFileRoot()
	// cause zhihu hot not register as corn job so load data by this time incase it has no data
	_, _ = LoadZhihuHot()
	zhihuFilePath, err := file_util.WriteJsonFile(z.GetData(), path, "zhihu", true)
	if err != nil {
		log.Errorf("Write zhihu hot daily record failed %s", err.Error())
	} else {
		cosPath := fmt.Sprintf("%s/%s", "zhihu", time.Now().Format("2006"))
		cosFileName := fmt.Sprintf("%s_%s.json", "zhihu", time.Now().Format("0102"))
		if err = file_util.TCCosUpload(cosPath, cosFileName, zhihuFilePath); err != nil {
			log.Errorf("Upload zhihu hot daily record to tencent cos failed %s", err.Error())
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
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/zhihu.json", path), &data); err != nil {
		log.Errorf("load zhihu hot daily record failed:%s", err.Error())
		data = make(map[string][]ZhihuHot)
	}
	z.data = data
}

func (d *zhihuHotDailyRecord) Backup() {
	d.RLock()
	defer d.RUnlock()
	//log.Infof("backup zhihu hot daily record")
	path := file_util.GetFileRoot()
	_, _ = file_util.WriteJsonFile(d.data, path, "zhihu", true)
}

type sentCache struct {
	data map[uint32]int64
	sync.RWMutex
}

func (z *sentCache) Add(title string) {
	z.Lock()
	defer z.Unlock()
	now := time.Now()
	if len(z.data) >= 3000 {
		for k, v := range z.data {
			if v < now.AddDate(0, 0, -3).Unix() {
				delete(z.data, k)
			}
		}

	}
	z.data[encrypt.HashStr(title)] = now.Unix()
}

func (z *sentCache) CheckSent(title string) bool {
	z.RLock()
	z.RUnlock()
	_, ok := z.data[encrypt.HashStr(title)]
	return ok
}

// 写入发送新闻记录文件
func (z *sentCache) Upload() {
	z.Lock()
	defer z.Unlock()
	path := file_util.GetFileRoot()
	// cause zhihu hot not register as corn job so load data by this time incase it has no data
	sentNewsFilePath, err := file_util.WriteJsonFile(z.data, path, "sentNews", true)
	if err != nil {
		log.Errorf("Write sent news record failed %s", err.Error())
	} else {
		cosPath := "cache"
		cosFileName := "sentNews"
		if err = file_util.TCCosUpload(cosPath, cosFileName, sentNewsFilePath); err != nil {
			log.Errorf("Upload sent news cache to tencent cos failed %s", err.Error())
		}
	}

}

// 加载发送新闻记录文件
func (z *sentCache) Load() {
	z.RWMutex = sync.RWMutex{}
	z.Lock()
	defer z.Unlock()
	path := file_util.GetFileRoot()
	data := make(map[uint32]int64)
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/sentNews.json", path), &data); err != nil {
		log.Errorf("laod sent news cache failed:%s", err.Error())
		data = make(map[uint32]int64)
	}
	z.data = data
}

func (d *sentCache) Backup() {
	d.RLock()
	defer d.RUnlock()
	//log.Infof("backup sent news cache")
	path := file_util.GetFileRoot()
	_, _ = file_util.WriteJsonFile(d.data, path, "sentNews", true)
}

// time.Now().Format("2006-01-02 15:04:05")
func UploadDailyRecord() {

	//写微博热搜当日文件
	WeiboHotDailyRecord.Upload()

	//写华尔街资讯当日文件
	WallStreetNewsDailyRecord.Upload()

	//写财新网新闻当日文件
	CaiXinNewsDailyRecord.Upload()

	//写36氪日榜当日文件
	D36krDailyRecord.Upload()

	//写知乎热榜当日文件
	ZhihuHotDailyRecord.Upload()

	//写新闻发送缓存文件
	SentRecord.Upload()

}

func Backup() {

	WeiboHotDailyRecord.Backup()

	WallStreetNewsDailyRecord.Backup()

	CaiXinNewsDailyRecord.Backup()

	D36krDailyRecord.Backup()

	ZhihuHotDailyRecord.Backup()

	SentRecord.Backup()
}

func init() {
	//加载微博热搜每日记录
	WeiboHotDailyRecord.Load()

	//加载华尔街每日记录
	WallStreetNewsDailyRecord.Load()

	//加载财新网每日记录
	CaiXinNewsDailyRecord.Load()

	//加载36氪每日记录
	D36krDailyRecord.Load()

	//加载知乎热榜每日记录
	ZhihuHotDailyRecord.Load()

	//加载新闻发送cache
	SentRecord.Load()

}
