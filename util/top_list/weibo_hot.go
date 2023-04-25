package top_list

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/http_util"
	"github.com/PuerkitoBio/goquery"
	"github.com/tristan-club/kit/log"
	"io"
	"net/http"
	"strings"
	"time"
)

type WeiboHot struct {
	Title string
	Hot   string
	Rank  int
	Url   string
}

// https://m.weibo.cn/api/container/getIndex?containerid=106003type%3D25%26t%3D3%26disable_hot%3D1%26filter_type%3Drealtimehot
// weiboHot analyze:https://m.s.weibo.com/topic/detail?q=%s
func ParseWeiboHotByApi() (map[string]interface{}, error) {
	url := "https://m.weibo.cn/api/container/getIndex?containerid=106003type%3D25%26t%3D3%26disable_hot%3D1%26filter_type%3Drealtimehot"
	var data map[string]interface{}
	err := http_util.GetJSON(url, nil, &data)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "request weibo api",
			"error":  err,
		}).Send()
	}
	return data, err
}

func LoadWeiboHot() ([]WeiboHot, error) {
	hotList, err := ParseWeiboHot()
	if WeiboHotDailyRecord == nil {
		WeiboHotDailyRecord = make(map[string][]WeiboHot)
	}
	WeiboHotDailyRecord[time.Now().Format("2006-01-02 15:04")] = hotList
	return hotList, err
}

func ParseWeiboHot() ([]WeiboHot, error) {
	url := "https://s.weibo.com/top/summary?cate=realtimehot"
	timeout := 120 * time.Second //超时时间2mine
	client := &http.Client{
		Timeout: timeout,
	}

	var hotList []WeiboHot

	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		return hotList, err
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Cookie", constant.COOKIE)
	//request.Header.add("Host", `wallstreetcn.com`)
	//request.Header.add("Referer", `https://wallstreetcn.com/`)
	res, err := client.Do(request)

	if err != nil {
		return hotList, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return hotList, err
	}

	titleCache := make(map[string]struct{})
	document.Find(".td-02").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()

		if boolUrl && strings.Contains(url, "weibo") {
			titleCache[text] = struct{}{}
			allData = append(allData, map[string]interface{}{"title": text, "url": fmt.Sprintf("https://s.weibo.com%s", url)})
		}
	})

	for _i, _data := range allData {
		hotList = append(hotList, WeiboHot{
			Title: _data["title"].(string),
			Rank:  _i + 1,
			Url:   _data["url"].(string),
		})
	}
	return hotList, err
}
