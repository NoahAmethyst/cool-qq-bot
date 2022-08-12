package weibo_hot

import (
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/opesun/goquery"
	"io/ioutil"
	"net/http"
)

type WeiboHot struct {
	Title string
	Hot   string
	Rank  int
}

func Summary() ([]WeiboHot, error) {
	var hotList []WeiboHot
	html, err := GetHTML(constant.WEIBO)
	if err != nil {
		return hotList, err
	}
	hotList, err = ParseWeiboHot(html)
	return hotList, err
}

func GetHTML(url string) (string, error) {
	var html string
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return html, err
	}
	req.Header.Set("User-Agent", constant.AGENT)
	req.Header.Set("Cookie", constant.COOKIE)
	res, err := client.Do(req)
	if err != nil {
		return html, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	html = string(body)
	return html, nil
}

func ParseWeiboHot(contents string) ([]WeiboHot, error) {
	html, err := goquery.ParseString(contents)
	if err != nil {
		return nil, err
	}

	var hotList []WeiboHot

	hotInfo := html.Find(".td-02")
	hotInfo.Find("a").Each(func(index int, element *goquery.Node) {
		for _, node := range element.Child {
			weiboHot := WeiboHot{
				Title: node.Data,
				Rank:  index + 1,
			}
			hotList = append(hotList, weiboHot)
		}
	})

	return hotList, err
}
