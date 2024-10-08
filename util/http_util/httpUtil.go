package http_util

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"

	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func Get(url string, headers map[string]string) ([]byte, error) {

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetJSON(url string, headers map[string]string, obj interface{}) error {
	b, err := Get(url, headers)

	if err != nil {
		return err
	}

	//var d interface{}
	err = json.Unmarshal(b, &obj)
	if err != nil {
		return err
	}
	return nil
}

func Post(url string, params interface{}, headers map[string]string) ([]byte, error) {

	bytesData, err := json.Marshal(params)

	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func PostJSON(url string, params interface{}, headers map[string]string, ret interface{}) error {

	if headers == nil {
		headers = make(map[string]string, 0)
	}
	headers["Content-Type"] = "application/json"
	b, err := Post(url, params, headers)

	if err != nil {
		log.Error(err)
		return err
	}

	if ret != nil {
		err = json.Unmarshal(b, ret)
	}

	return err
}

func PostForm(thisUrl string, params map[string]string, headers map[string]string, result interface{}) error {

	headers["Content-Type"] = "application/x-www-form-urlencoded"
	values := url.Values{}
	a := []string{}
	for k, v := range params {
		a = append(a, v)
		values[k] = a
		a = []string{}
	}

	res, err := http.PostForm(thisUrl, values)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	log.Infof("%+v", map[string]interface{}{
		"action":   "post form",
		"url":      thisUrl,
		"response": result,
	})

	return err
}
