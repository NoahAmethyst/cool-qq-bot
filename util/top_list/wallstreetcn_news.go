package top_list

import (
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/encrypt"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"os"
	"time"
)

type WallStreetNews struct {
	Title   string
	Content string
	Url     string
}

func LoadWallStreetNews() ([]WallStreetNews, error) {
	data, err := ParseWallStreetNews()
	currData := make([]WallStreetNews, 0, 50)

	if recordData := WallStreetNewsDailyRecord.GetData(); len(recordData) > 0 {
		restoredSet := make(map[uint32]struct{})

		for _, news := range recordData {
			for _, _existNew := range news {
				restoredSet[encrypt.HashStr(_existNew.Title)] = struct{}{}
			}
		}
		for _, _data := range data {
			if _, ok := restoredSet[encrypt.HashStr(_data.Title)]; !ok {
				currData = append(currData, _data)
			}
		}
	} else {
		currData = data
	}

	if len(currData) > 0 {
		WallStreetNewsDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), currData)
	}

	go func() {
		path := os.Getenv(constant.FILE_ROOT)
		if len(path) == 0 {
			path = "/tmp"
		}
		_, _ = file_util.WriteJsonFile(WallStreetNewsDailyRecord.GetData(), path, "wallstreet_news", true)
	}()

	return data, err
}

func ParseWallStreetNews() ([]WallStreetNews, error) {
	url := "https://wallstreetcn.com/news/global"
	timeout := 120 * time.Second //超时时间2mine
	client := &http.Client{
		Timeout: timeout,
	}

	var data []WallStreetNews
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

	data = make([]WallStreetNews, 0, 11)

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
