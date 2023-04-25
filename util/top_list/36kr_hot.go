package top_list

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"time"
)

type Data36krHot struct {
	Title string
	Url   string
	Rank  int
}

func Load36krHot() ([]Data36krHot, error) {
	data, err := Parse36krHot()
	if Data36krDailyRecord == nil {
		Data36krDailyRecord = make(map[string][]Data36krHot)
	}
	Data36krDailyRecord[time.Now().Format("2006-01-02 15:04")] = data
	return data, err
}

func Parse36krHot() ([]Data36krHot, error) {
	var data []Data36krHot
	url := "https://36kr.com/"
	timeout := 120 * time.Second //超时时间2mine
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		return data, err
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `36kr.com`)
	request.Header.Add("Referer", `https://36kr.com/`)
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
	document.Find(".hotlist-item-toptwo").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := selection.Find("a p").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://36kr.com" + url})
		}
	})
	document.Find(".hotlist-item-other-info").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://36kr.com" + url})
		}
	})

	data = make([]Data36krHot, len(allData))

	for _i, _data := range allData {
		data[_i] = Data36krHot{
			Title: _data["title"].(string),
			Url:   _data["url"].(string),
			Rank:  _i + 1,
		}
	}
	return data, nil
}
