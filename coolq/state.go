package coolq

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/encrypt"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/tristan-club/kit/log"
	"os"
	"sync"
	"time"
)

var once sync.Once

type State struct {
	wallstreetSentNews     *wallStreetSentNews
	assistantModel         *assistantModel
	groupDialogueSession   *aiAssistantSession
	privateDialogueSession *aiAssistantSession
}

// wallStreetSentNews 华尔街日报发送记录
type wallStreetSentNews struct {
	sync.RWMutex
	SentList map[int64]map[uint32]time.Time
}

// assistantModel Todo 用户chatgpt模型选择
type assistantModel struct {
}

// aiAssistantSession record ai assistant chat conversation to maintain the context of the conversation
type aiAssistantSession struct {
	sessionChan map[int64]chan string
	parentId    map[int64]string
	sync.RWMutex
}

func (s *wallStreetSentNews) add(group int64, title string) {
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	if _, ok := s.SentList[group]; !ok {
		s.SentList[group] = map[uint32]time.Time{
			encrypt.HashStr(title): now,
		}
	} else {
		s.SentList[group][encrypt.HashStr(title)] = now
	}
	if len(s.SentList[group]) > 200 {
		for _titleHash, _createdAt := range s.SentList[group] {
			if now.Sub(_createdAt) > 24*time.Hour {
				delete(s.SentList[group], _titleHash)
			}
		}
	}
}

func (s *wallStreetSentNews) checkSent(group int64, title string) bool {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.SentList[group]; !ok {
		return ok
	} else {
		_, ok := v[encrypt.HashStr(title)]
		return ok
	}

}

func (s *wallStreetSentNews) SaveCache() {
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

func (s *aiAssistantSession) putParentMsgId(uid int64, parentMsgId string) {
	s.Lock()
	defer s.Unlock()
	if s.sessionChan[uid] == nil {
		s.sessionChan[uid] = make(chan string)
		go func(int64) {
			for {
				select {
				case id := <-s.sessionChan[uid]:
					s.setParentMsgId(uid, id)
				case <-time.After(time.Minute * 10):
					s.delParentId(uid)
				}
			}
		}(uid)
	}
	s.sessionChan[uid] <- parentMsgId
}

func (s *aiAssistantSession) getParentMsgId(uid int64) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.parentId[uid]
	return v, ok
}

func (s *aiAssistantSession) setParentMsgId(uid int64, parentMsgId string) {
	s.Lock()
	defer s.Unlock()
	s.parentId[uid] = parentMsgId
}

func (s *aiAssistantSession) delParentId(uid int64) {
	s.Lock()
	defer s.Unlock()
	delete(s.parentId, uid)
}

func (b *CQBot) initState() {
	once.Do(func() {
		if b.state == nil {
			b.state = &State{
				wallstreetSentNews: initWallStreetSentNews(),
				assistantModel:     &assistantModel{},
				groupDialogueSession: &aiAssistantSession{
					sessionChan: map[int64]chan string{},
					parentId:    map[int64]string{},
					RWMutex:     sync.RWMutex{},
				},
				privateDialogueSession: &aiAssistantSession{
					sessionChan: map[int64]chan string{},
					parentId:    map[int64]string{},
					RWMutex:     sync.RWMutex{},
				},
			}
		}
	})

}

func initWallStreetSentNews() *wallStreetSentNews {
	SentNews := wallStreetSentNews{
		SentList: map[int64]map[uint32]time.Time{},
		RWMutex:  sync.RWMutex{},
	}
	data := make(map[int64]map[uint32]time.Time)
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
