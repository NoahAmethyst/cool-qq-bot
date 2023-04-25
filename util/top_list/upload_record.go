package top_list

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/tristan-club/kit/log"
	"os"
	"time"
)

var Data36krDailyRecord map[string][]Data36krHot
var WallStreetNewsDailyRecord map[string][]WallStreetNews
var WeiboHotDailyRecord map[string][]WeiboHot

// time.Now().Format("2006-01-02 15:04:05")
func UploadDailyRecord() {
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}

	//写微博热搜当日文件
	{
		weiboFilePath, err := file_util.WriteJsonFile(WeiboHotDailyRecord, path, "weibo_hot", true)
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
				WeiboHotDailyRecord = nil
			}
		}
	}

	//写华尔街资讯当日文件
	{
		wallStreetFilePath, err := file_util.WriteJsonFile(WeiboHotDailyRecord, path, "wallstreet_news", true)
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
				WallStreetNewsDailyRecord = nil
			}
		}
	}

	//写36氪日榜当日文件
	{
		d36krFilePath, err := file_util.WriteJsonFile(WeiboHotDailyRecord, path, "36kr", true)
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
				Data36krDailyRecord = nil
			}
		}
	}

}
