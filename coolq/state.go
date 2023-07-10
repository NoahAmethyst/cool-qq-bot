package coolq

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/ai_util"
	"github.com/Mrs4s/go-cqhttp/util/encrypt"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	bingchat_api "github.com/NoahAmethyst/bingchat-api"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
	"strconv"

	"os"
	"sync"
	"time"
)

var once sync.Once

type State struct {
	owner                  int64
	reportState            *reportState
	wallstreetSentNews     *wallStreetSentNews
	assistantModel         *assistantModel
	groupDialogueSession   *AiAssistantSession
	privateDialogueSession *AiAssistantSession
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
	path := file_util.GetFileRoot()
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
	path := file_util.GetFileRoot()
	_, err := file_util.WriteJsonFile(a.selectedModel, path, "assistant_model", false)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "save assistant_model to file",
			"error":  err,
		}).Send()
	} else {
		_ = file_util.TCCosUpload("cache", "assistant_model.json", fmt.Sprintf("%s/%s", path, "assistant_model.json"))
	}
}

func (a *assistantModel) getModel(uid int64) ai_util.ChatModel {
	a.RLock()
	defer a.RUnlock()
	return a.selectedModel[uid]
}

// AiAssistantSession record ai assistant chat conversation to maintain the context of the conversation
type AiAssistantSession struct {
	//chat assistant
	assistantChan map[int64]chan string
	parentId      map[int64]string
	//bingchat
	bingChan     map[int64]chan struct{}
	conversation map[int64]bingchat_api.IBingChat
	//chatgpt
	chatgptChan map[int64]chan struct{}
	ctx         map[int64][]openai.ChatCompletionMessage

	sync.RWMutex
}

func (s *AiAssistantSession) putParentMsgId(uid int64, parentMsgId string) {
	s.Lock()
	defer s.Unlock()
	if s.assistantChan[uid] == nil {
		s.assistantChan[uid] = make(chan string)
		go func(int64) {
			for {
				select {
				case id := <-s.assistantChan[uid]:
					s.setParentMsgId(uid, id)
				case <-time.After(time.Minute * 10):
					s.delParentId(uid)
				}
			}
		}(uid)
	}
	s.assistantChan[uid] <- parentMsgId
}

func (s *AiAssistantSession) getParentMsgId(uid int64) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.parentId[uid]
	return v, ok
}

func (s *AiAssistantSession) setParentMsgId(uid int64, parentMsgId string) {
	s.Lock()
	defer s.Unlock()
	s.parentId[uid] = parentMsgId
}

func (s *AiAssistantSession) delParentId(uid int64) {
	s.Lock()
	defer s.Unlock()
	delete(s.parentId, uid)
}

func (s *AiAssistantSession) putConversation(uid int64, conversation bingchat_api.IBingChat) {
	s.Lock()
	defer s.Unlock()
	if s.conversation[uid] == nil {
		s.conversation[uid] = conversation
	}
	if s.bingChan[uid] == nil {
		s.bingChan[uid] = make(chan struct{})
		go func(int64) {
			for {
				select {
				case <-s.bingChan[uid]:

				case <-time.After(time.Minute * 5):
					s.closeConversation(uid)
				}
			}
		}(uid)
	}
	s.bingChan[uid] <- struct{}{}
}

func (s *AiAssistantSession) getConversation(uid int64) bingchat_api.IBingChat {
	s.RLock()
	defer s.RUnlock()
	return s.conversation[uid]
}

func (s *AiAssistantSession) closeConversation(uid int64) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.conversation[uid]; ok {
		s.conversation[uid].Close()
	}

	delete(s.conversation, uid)
}

func (s *AiAssistantSession) putCtx(uid int64, msg, resp string) {
	s.Lock()
	defer s.Unlock()
	if s.ctx[uid] == nil {
		s.ctx[uid] = make([]openai.ChatCompletionMessage, 0, 8)
	}
	s.ctx[uid] = append(s.ctx[uid], []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: msg,
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: resp,
		},
	}...)

	if s.chatgptChan[uid] == nil {
		s.chatgptChan[uid] = make(chan struct{})
		go func(int64) {
			for {
				select {
				case <-s.chatgptChan[uid]:

				case <-time.After(time.Minute * 10):
					s.clearCtx(uid)
				}
			}
		}(uid)
	}
	s.chatgptChan[uid] <- struct{}{}
}

func (s *AiAssistantSession) getCtx(uid int64) []openai.ChatCompletionMessage {
	s.RLock()
	defer s.RUnlock()
	ctx := make([]openai.ChatCompletionMessage, len(s.ctx[uid]))
	copy(ctx, s.ctx[uid])
	return ctx
}

func (s *AiAssistantSession) clearCtx(uid int64) {
	s.Lock()
	defer s.Unlock()
	delete(s.ctx, uid)
}

func (bot *CQBot) initState() {
	once.Do(func() {
		var owner int64
		if len(os.Getenv(constant.OWNER)) > 0 {
			owner, _ = strconv.ParseInt(os.Getenv(constant.OWNER), 10, 64)
		}
		if bot.state == nil {
			bot.state = &State{
				owner:              owner,
				reportState:        initReportState(),
				wallstreetSentNews: initWallStreetSentNews(),
				assistantModel:     initAssistantModel(),
				groupDialogueSession: &AiAssistantSession{
					assistantChan: map[int64]chan string{},
					parentId:      map[int64]string{},
					bingChan:      map[int64]chan struct{}{},
					conversation:  map[int64]bingchat_api.IBingChat{},
					chatgptChan:   map[int64]chan struct{}{},
					ctx:           map[int64][]openai.ChatCompletionMessage{},
					RWMutex:       sync.RWMutex{},
				},
				privateDialogueSession: &AiAssistantSession{
					assistantChan: map[int64]chan string{},
					parentId:      map[int64]string{},
					bingChan:      map[int64]chan struct{}{},
					conversation:  map[int64]bingchat_api.IBingChat{},
					chatgptChan:   map[int64]chan struct{}{},
					ctx:           map[int64][]openai.ChatCompletionMessage{},
					RWMutex:       sync.RWMutex{},
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
	path := file_util.GetFileRoot()
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
	path := file_util.GetFileRoot()
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

func initAssistantModel() *assistantModel {
	_assistantModel := assistantModel{
		selectedModel: map[int64]ai_util.ChatModel{},
		RWMutex:       sync.RWMutex{},
	}
	data := make(map[int64]ai_util.ChatModel)
	path := file_util.GetFileRoot()
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/assistant_model.json", path), &data); err != nil {
		log.Info().Fields(map[string]interface{}{
			"action": "retry load assistant_model json from tencent cos",
		}).Send()
		_err := file_util.TCCosDownload("cache", "assistant_model.json", fmt.Sprintf("%s/%s", path, "assistant_model.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/assistant_model.json", path), &data)
		}
	}
	if len(data) > 0 {
		_assistantModel.selectedModel = data
	}
	return &_assistantModel
}
