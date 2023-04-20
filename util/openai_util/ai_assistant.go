package openai_util

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	aiassistantUrl = "https://aigc.zhyjor.com/ai-assistant/api/conversation-private"
)

// Request represents the request body
type AIAssistantReq struct {
	APIKey   string `json:"apiKey"`
	Question string `json:"question"`
	Options  struct {
		ParentMessageId string `json:"parentMessageId"`
		SystemMessage   string `json:"systemMessage"`
	} `json:"options"`
}

// Response represents the response body
type AIAssistantResp struct {
	Role            string `json:"role"`
	ID              string `json:"id"`
	ParentMessageID string `json:"parentMessageId"`
	Text            string `json:"text"`
	Detail          struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int    `json:"created"`
		Model   string `json:"model"`
		Usage   struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
			Index        int    `json:"index"`
		} `json:"choices"`
	} `json:"detail"`
}

func AskAIAssistant(content string, option ...string) (*AIAssistantResp, error) {
	// Create request body
	var requestBody AIAssistantReq
	if len(option) == 0 {
		requestBody = AIAssistantReq{
			APIKey:   "ai-assistant",
			Question: content,
			Options: struct {
				ParentMessageId string `json:"parentMessageId"`
				SystemMessage   string `json:"systemMessage"`
			}{
				SystemMessage: "You are ChatGPT, a large language model trained by OpenAI.",
			},
		}
	} else {
		requestBody = AIAssistantReq{
			APIKey:   "ai-assistant",
			Question: content,
			Options: struct {
				ParentMessageId string `json:"parentMessageId"`
				SystemMessage   string `json:"systemMessage"`
			}{
				ParentMessageId: option[0],
				SystemMessage:   "You are ChatGPT, a large language model trained by OpenAI.",
			},
		}
	}

	// Marshal request body to JSON
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", aiassistantUrl, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		panic(err)
	}

	// Set request headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Origin", "https://aigc.zhyjor.com")
	req.Header.Set("Referer", "https://aigc.zhyjor.com/ai-assistant")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Google Chrome\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"macOS\"")

	// Send HTTP request
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read response body
	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response body to struct
	var response AIAssistantResp
	err = json.Unmarshal(responseBodyBytes, &response)
	if err != nil {
		panic(err)
	}

	return &response, nil

}
