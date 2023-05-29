package coolq

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/ai_util"
	"github.com/Mrs4s/go-cqhttp/util/encrypt"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/rs/zerolog/log"

	"os"
	"sync"
	"time"
)

var once sync.Once

type State struct {
	reportState            *reportState
	wallstreetSentNews     *wallStreetSentNews
	assistantModel         *assistantModel
	groupDialogueSession   *aiAssistantSession
	privateDialogueSession *aiAssistantSession
}

type reportState struct {
	sync.RWMutex
	groups   map[int64]struct{}
	privates map[int64]struct{}
}

func (r *reportState) saveCache() {
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}

	//group
	{
		_, err := file_util.WriteJsonFile(r.groups, path, "reportGroupState", false)
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "save group report state to file",
				"error":  err,
			}).Send()
		} else {
			_ = file_util.TCCosUpload("cache", "reportGroupState.json", fmt.Sprintf("%s/%s", path, "reportGroupState.json"))
		}
	}

	//private
	{
		_, err := file_util.WriteJsonFile(r.privates, path, "reportPrivateState", false)
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "save private report state to file",
				"error":  err,
			}).Send()
		} else {
			_ = file_util.TCCosUpload("cache", "reportPrivateState.json", fmt.Sprintf("%s/%s", path, "reportPrivateState.json"))
		}
	}

}

func (r *reportState) add(id int64, isGroup bool) {
	r.Lock()
	defer r.Unlock()
	if isGroup {
		r.groups[id] = struct{}{}
	} else {
		r.privates[id] = struct{}{}
	}
	r.saveCache()
}

func (r *reportState) del(id int64, isGroup bool) {
	r.Lock()
	defer r.Unlock()
	if isGroup {
		delete(r.groups, id)
	} else {
		delete(r.privates, id)
	}
	r.saveCache()
}

func (r *reportState) exist(id int64, isGroup bool) bool {
	r.RLock()
	defer r.RUnlock()
	var ok bool
	if isGroup {
		_, ok = r.groups[id]
	} else {
		_, ok = r.privates[id]
	}
	return ok
}

func (r *reportState) getReportList(isGroup bool) map[int64]struct{} {
	r.RLock()
	defer r.RUnlock()
	data := make(map[int64]struct{})
	if isGroup {
		for k, v := range r.groups {
			data[k] = v
		}
	} else {
		for k, v := range r.privates {
			data[k] = v
		}
	}

	return data
}

// wallStreetSentNews 华尔街日报发送记录
type wallStreetSentNews struct {
	sync.RWMutex
	SentList map[int64]map[uint32]time.Time
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
	if len(s.SentList[group]) > 20 {
		for _titleHash, _createdAt := range s.SentList[group] {
			if now.Sub(_createdAt) > 12*time.Hour {
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

// assistantModel  用户chatgpt模型选择
type assistantModel struct {
	selectedModel map[int64]ai_util.ChatModel
	sync.RWMutex
}

func (a *assistantModel) setModel(uid int64, model ai_util.ChatModel) {
	a.Lock()
	defer a.Unlock()
	a.selectedModel[uid] = model
}

func (a *assistantModel) getModel(uid int64) ai_util.ChatModel {
	a.RLock()
	defer a.RUnlock()
	return a.selectedModel[uid]
}

// aiAssistantSession record ai assistant chat conversation to maintain the context of the conversation
type aiAssistantSession struct {
	sessionChan map[int64]chan string
	parentId    map[int64]string
	sync.RWMutex
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

func (bot *CQBot) initState() {
	once.Do(func() {
		if bot.state == nil {
			bot.state = &State{
				reportState:        initReportState(),
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

func initReportState() *reportState {
	_reportState := reportState{
		groups:   map[int64]struct{}{},
		privates: map[int64]struct{}{},
		RWMutex:  sync.RWMutex{},
	}
	groupData := make(map[int64]struct{})
	privateData := make(map[int64]struct{})
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/reportGroupState.json", path), &groupData); err != nil {
		log.Info().Fields(map[string]interface{}{
			"action": "retry load wallstreet json from tencent cos",
		}).Send()
		_err := file_util.TCCosDownload("cache", "reportGroupState.json", fmt.Sprintf("%s/%s", path, "reportGroupState.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/reportGroupState.json", path), &groupData)
		}
	}

	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/reportPrivateState.json", path), &privateData); err != nil {
		log.Info().Fields(map[string]interface{}{
			"action": "retry load wallstreet json from tencent cos",
		}).Send()
		_err := file_util.TCCosDownload("cache", "reportPrivateState.json", fmt.Sprintf("%s/%s", path, "reportPrivateState.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/reportPrivateState.json", path), &privateData)
		}
	}
	if len(groupData) > 0 {
		_reportState.groups = groupData
	}

	if len(privateData) > 0 {
		_reportState.privates = privateData
	}
	return &_reportState
}

func initWallStreetSentNews() *wallStreetSentNews {
	sentNews := wallStreetSentNews{
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
		sentNews.SentList = data
	}
	return &sentNews
}
