package top_list

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/PuerkitoBio/goquery"
	"github.com/tristan-club/kit/log"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type WallStreetNews struct {
	Title   string
	Content string
	Url     string
}

var SentNews SentNewsRecord

type SentNewsRecord struct {
	SentList map[string]time.Time
	sync.RWMutex
}

func LoadWallStreetNews() ([]WallStreetNews, error) {
	data, err := ParseWallStreetNews()
	var readyData []WallStreetNews
	for _, _data := range data {
		if !SentNews.checkSent(_data.Title) {
			readyData = append(readyData, _data)
			SentNews.add(_data.Title)
		}
	}

	if len(readyData) == 0 {
		log.Warn().Msgf("华尔街见闻：没有最新资讯，爬取资讯数量:%d", len(data))
	} else {
		SentNews.SaveCache()
	}

	return readyData, err
}

func ParseWallStreetNews() ([]WallStreetNews, error) {
	url := "https://wallstreetcn.com/news/global"
	timeout := 120 * time.Second //超时时间2mine
	client := &http.Client{
		Timeout: timeout,
	}

	var data []WallStreetNews
	{
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		return data, err
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	//request.Header.add("Host", `wallstreetcn.com`)
	//request.Header.add("Referer", `https://wallstreetcn.com/`)
	res, err := client.Do(request)

	if err != nil {
		return data, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return data, err
	}
	document.Find(".article-entry").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := selection.Find("span").Text()
		content := selection.Find(".content").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "content": content, "url": url})
		}
	})

	for _i, _data := range allData {
		if _i > 10 {
			break
		}
		data = append(data, WallStreetNews{
			Title:   _data["title"].(string),
			Content: _data["content"].(string),
			Url:     _data["url"].(string),
		})
	}
	return data, nil
}

func (s *SentNewsRecord) add(title string) {
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	s.SentList[title] = now
	if len(s.SentList) > 200 {
		for _title, _createdAt := range s.SentList {
			if now.Sub(_createdAt) > 3*time.Hour {
				delete(s.SentList, _title)
			}
		}
	}
}

func (s *SentNewsRecord) checkSent(title string) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.SentList[title]
	return ok
}

func (s *SentNewsRecord) SaveCache() {
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
	}
}

func init() {
	SentNews = SentNewsRecord{
		SentList: map[string]time.Time{},
		RWMutex:  sync.RWMutex{},
	}
	data := make(map[string]time.Time)
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}
	_ = file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data)
	if len(data) > 0 {
		SentNews.SentList = data
	}
}
