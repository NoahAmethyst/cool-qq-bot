package coolq

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/tristan-club/kit/log"
	"os"
	"sync"
	"time"
)

type State struct {
	wallstreetSentNews *WallStreetSentNews
	assistantModel     *AssistantModel
}

// WallStreetSentNews 华尔街日报发送记录
type WallStreetSentNews struct {
	sync.RWMutex
	SentList map[int64]map[string]time.Time
}

// AssistantModel Todo 用户chatgpt模型选择
type AssistantModel struct {
}

func (s *WallStreetSentNews) add(group int64, title string) {
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	if _, ok := s.SentList[group]; !ok {
		s.SentList[group] = map[string]time.Time{
			title: now,
		}
	} else {
		s.SentList[group][title] = now
	}
	if len(s.SentList[group]) > 200 {
		for _title, _createdAt := range s.SentList[group] {
			if now.Sub(_createdAt) > 24*time.Hour {
				delete(s.SentList[group], _title)
			}
		}
	}
}

func (s *WallStreetSentNews) checkSent(group int64, title string) bool {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.SentList[group]; !ok {
		return ok
	} else {
		_, ok := v[title]
		return ok
	}

}

func (s *WallStreetSentNews) SaveCache() {
	s.RLock()
	defer s.RUnlock()
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}
	_, err := file_util.WriteJsonFile(s.SentList, path, "wallStreetCache", false)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "save wall street news to file",
			"error":  err,
		}).Send()
	} else {
		_ = file_util.TCCosUpload("cache", "wallStreetCache.json", fmt.Sprintf("%s/%s", path, "wallStreetCache.json"))
	}
}

func initWallStreetSentNews() *WallStreetSentNews {
	SentNews := WallStreetSentNews{
		SentList: map[int64]map[string]time.Time{},
		RWMutex:  sync.RWMutex{},
	}
	data := make(map[int64]map[string]time.Time)
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data); err != nil {
		log.Info().Fields(map[string]interface{}{
			"action": "retry load wallstreet json from tencent cos",
		}).Send()
		_err := file_util.TCCosDownload("cache", "wallStreetCache.json", fmt.Sprintf("%s/%s", path, "wallStreetCache.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data)
		}
	}
	if len(data) > 0 {
		SentNews.SentList = data
	}
	return &SentNews
}
