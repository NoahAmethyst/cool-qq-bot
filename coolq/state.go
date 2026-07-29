package coolq

import (
	"fmt"
	"strconv"

	"github.com/NoahAmethyst/go-cqhttp/constant"
	"github.com/NoahAmethyst/go-cqhttp/util/encrypt"
	"github.com/NoahAmethyst/go-cqhttp/util/file_util"
	log "github.com/sirupsen/logrus"

	"os"
	"sync"
	"time"
)

var once sync.Once

type State struct {
	owner       int64
	reportState *reportState
	sentNews    *sentNews
	globalState *globalState
}

type globalState struct {
	Data map[string]string
	sync.RWMutex
}

func (s *globalState) Set(k string, v string) {
	s.Lock()
	defer s.Unlock()
	if len(s.Data) == 0 {
		s.Data = make(map[string]string)
	}
	s.Data[k] = v

}

func (s *globalState) GetData(k string) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.Data[k]
	return v, ok

}

func (s *globalState) SaveCache() {
	s.RLock()
	defer s.RUnlock()
	path := file_util.GetFileRoot()
	_, err := file_util.WriteJsonFile(s.Data, path, "global_state", false)
	if err != nil {
		log.Errorf("save wall street news to file faild:%s", err.Error())
	} else {
		_ = file_util.TCCosUpload("cache", "global_state.json", fmt.Sprintf("%s/%s", path, "global_state.json"))
	}
}

func (s *State) SaveCache() {
	log.Infof("save bot state cache")
	s.reportState.saveCache()
	s.sentNews.SaveCache()
	s.globalState.SaveCache()
}

type reportState struct {
	sync.RWMutex
	groups   map[int64]struct{}
	privates map[int64]struct{}
}

func (r *reportState) saveCache() {
	path := file_util.GetFileRoot()
	//group
	{
		_, err := file_util.WriteJsonFile(r.groups, path, "reportGroupState", false)
		if err != nil {
			log.Errorf("save group report state to file faield:%s", err.Error())
		} else {
			_ = file_util.TCCosUpload("cache", "reportGroupState.json", fmt.Sprintf("%s/%s", path, "reportGroupState.json"))
		}
	}

	//private
	{
		_, err := file_util.WriteJsonFile(r.privates, path, "reportPrivateState", false)
		if err != nil {
			log.Errorf("save private report state to file faield:%s", err.Error())
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

func (r *reportState) getReportList(isGroup bool) []int64 {
	r.RLock()

	defer r.RUnlock()
	groupIds := make([]int64, 0, 4)

	if isGroup {
		for k := range r.groups {
			groupIds = append(groupIds, k)
		}
	} else {
		for k := range r.privates {
			groupIds = append(groupIds, k)
		}
	}

	return groupIds
}

// sentNews 华尔街日报发送记录
type sentNews struct {
	sync.RWMutex
	SentList map[int64]map[uint32]time.Time
}

func (s *sentNews) add(group int64, title string) {
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
	if len(s.SentList[group]) > 3600 {
		for _titleHash, _createdAt := range s.SentList[group] {
			if now.Sub(_createdAt) > 3*24*time.Hour {
				delete(s.SentList[group], _titleHash)
			}
		}
	}
}

func (s *sentNews) checkSent(group int64, title string) bool {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.SentList[group]; !ok {
		return ok
	} else {
		_, ok := v[encrypt.HashStr(title)]
		return ok
	}

}

func (s *sentNews) SaveCache() {
	s.RLock()
	defer s.RUnlock()
	path := file_util.GetFileRoot()
	_, err := file_util.WriteJsonFile(s.SentList, path, "wallStreetCache", false)
	if err != nil {
		log.Errorf("save wall street news to file faild:%s", err.Error())
	} else {
		_ = file_util.TCCosUpload("cache", "wallStreetCache.json", fmt.Sprintf("%s/%s", path, "wallStreetCache.json"))
	}
}

func (bot *CQBot) initState() {
	once.Do(func() {
		var owner int64
		if len(os.Getenv(constant.OWNER)) > 0 {
			owner, _ = strconv.ParseInt(os.Getenv(constant.OWNER), 10, 64)
		}
		if bot.state == nil {
			bot.state = &State{
				owner:       owner,
				reportState: initReportState(),
				sentNews:    initSentNews(),
				globalState: initGlobalState(),
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
	path := file_util.GetFileRoot()
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/reportGroupState.json", path), &groupData); err != nil {
		log.Info("retry load wallstreet json from tencent cos")
		_err := file_util.TCCosDownload("cache", "reportGroupState.json", fmt.Sprintf("%s/%s", path, "reportGroupState.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/reportGroupState.json", path), &groupData)
		}
	}

	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/reportPrivateState.json", path), &privateData); err != nil {
		log.Info("retry load wallstreet json from tencent cos")
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

func initSentNews() *sentNews {
	_sentNews := sentNews{
		SentList: map[int64]map[uint32]time.Time{},
		RWMutex:  sync.RWMutex{},
	}
	data := make(map[int64]map[uint32]time.Time)
	path := file_util.GetFileRoot()
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data); err != nil {
		log.Info("retry load wallstreet json from tencent cos")
		_err := file_util.TCCosDownload("cache", "wallStreetCache.json", fmt.Sprintf("%s/%s", path, "wallStreetCache.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data)
		}
	}
	if len(data) > 0 {
		_sentNews.SentList = data
	}
	return &_sentNews
}

func initGlobalState() *globalState {
	_globalState := globalState{
		Data:    map[string]string{},
		RWMutex: sync.RWMutex{},
	}
	data := make(map[string]string)
	path := file_util.GetFileRoot()
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/global_state.json", path), &data); err != nil {
		log.Info("retry load global_state json from tencent cos")
		_err := file_util.TCCosDownload("cache", "global_state.json", fmt.Sprintf("%s/%s", path, "global_state.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/global_state.json", path), &data)
		}
	}
	if len(data) > 0 {
		_globalState.Data = data
	}
	return &_globalState
}
