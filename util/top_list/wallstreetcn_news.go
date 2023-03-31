package top_list

import (
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
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
		if !SentNews.CheckSent(_data.Title) {
			readyData = append(readyData, _data)
			SentNews.Add(_data.Title)
		}
	}

	if len(readyData) == 0 {
		log.Warn("华尔街见闻：没有最新资讯，爬取资讯数量:%d", len(data))
	}
	return readyData, err
}

func ParseWallStreetNews() ([]WallStreetNews, error) {
	url := "https://wallstreetcn.com/news/global"
	timeout := time.Duration(5 * time.Second) //超时时间5s
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
	//request.Header.Add("Host", `wallstreetcn.com`)
	//request.Header.Add("Referer", `https://wallstreetcn.com/`)
	res, err := client.Do(request)

	if err != nil {
		return data, err
	}
	defer res.Body.Close()
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
		if _i > 6 {
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

func (s *SentNewsRecord) Add(title string) {
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	s.SentList[title] = now
	if len(s.SentList) > 100 {
		for _title, _createdAt := range s.SentList {
			if now.Sub(_createdAt) > 72*time.Hour {
				delete(s.SentList, _title)
			}
		}
	}
}

func (s *SentNewsRecord) CheckSent(title string) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.SentList[title]
	return ok
}

func init() {
	SentNews = SentNewsRecord{
		SentList: map[string]time.Time{},
		RWMutex:  sync.RWMutex{},
	}
}
