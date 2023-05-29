package top_list

import (
	"encoding/json"
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ZhihuHot struct {
	Title   string
	Url     string
	Excerpt string
	Rank    int
	Created int64
}

type zhihuHotList struct {
	Data []struct {
		Type      string `json:"type"`
		StyleType string `json:"style_type"`
		Id        string `json:"id"`
		CardId    string `json:"card_id"`
		Target    struct {
			Id            int    `json:"id"`
			Title         string `json:"title"`
			Url           string `json:"url"`
			Type          string `json:"type"`
			Created       int    `json:"created"`
			AnswerCount   int    `json:"answer_count"`
			FollowerCount int    `json:"follower_count"`
			Author        struct {
				Type      string `json:"type"`
				UserType  string `json:"user_type"`
				Id        string `json:"id"`
				UrlToken  string `json:"url_token"`
				Url       string `json:"url"`
				Name      string `json:"name"`
				Headline  string `json:"headline"`
				AvatarUrl string `json:"avatar_url"`
			} `json:"author"`
			BoundTopicIds []int  `json:"bound_topic_ids"`
			CommentCount  int    `json:"comment_count"`
			IsFollowing   bool   `json:"is_following"`
			Excerpt       string `json:"excerpt"`
		} `json:"target"`
		AttachedInfo string `json:"attached_info"`
		DetailText   string `json:"detail_text"`
		Trend        int    `json:"trend"`
		Debut        bool   `json:"debut"`
		Children     []struct {
			Type      string `json:"type"`
			Thumbnail string `json:"thumbnail"`
		} `json:"children"`
		CardLabel struct {
			Type      string `json:"type"`
			Icon      string `json:"icon"`
			NightIcon string `json:"night_icon"`
		} `json:"card_label,omitempty"`
	} `json:"data"`
	Paging struct {
		IsEnd    bool   `json:"is_end"`
		Next     string `json:"next"`
		Previous string `json:"previous"`
	} `json:"paging"`
	FreshText      string      `json:"fresh_text"`
	DisplayNum     int         `json:"display_num"`
	DisplayFirst   interface{} `json:"display_first"`
	FbBillMainRise int         `json:"fb_bill_main_rise"`
}

func LoadZhihuHot() ([]ZhihuHot, error) {
	hotList, err := parseZhihuHot()

	ZhihuHotDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), hotList)

	go func() {
		path := os.Getenv(constant.FILE_ROOT)
		if len(path) == 0 {
			path = "/tmp"
		}
		_, _ = file_util.WriteJsonFile(ZhihuHotDailyRecord.GetData(), path, "zhihu", true)
	}()

	//only return top 10 of the hot list
	return hotList[:10], err
}

func parseZhihuHot() ([]ZhihuHot, error) {
	hostList := make([]ZhihuHot, 0, 100)
	var respData zhihuHotList
	resp, err := http.Get("https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=100")
	if err != nil {
		return hostList, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &respData)
		for _index, _data := range respData.Data {
			currUrl := strings.ReplaceAll(_data.Target.Url, "api.", "")
			currUrl = strings.ReplaceAll(currUrl, "questions", "question")
			hostList = append(hostList, ZhihuHot{
				Title:   _data.Target.Title,
				Url:     currUrl,
				Excerpt: _data.Target.Excerpt,
				Created: int64(_data.Target.Created),
				Rank:    _index + 1,
			})
		}
	} else {
		err = errors.New(fmt.Sprintf("error:%d", resp.StatusCode))
	}
	return hostList, err
}
