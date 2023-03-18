package chatgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"wallcrow/lib"
	"wallcrow/lib/logger"
	"wallcrow/model"
)

type Response struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`

	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    model.Role `json:"role"`
			Content string     `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`

	Error *struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    *string `json:"code"`
	} `json:"error"`

	Status string `json:"status"`
}

/**
```shell
curl https://api.openai.com/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_API_KEY' \
  -d '{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hello!"}]
}'

```
*/

func ask(conv *model.Conversation, apikey string) (*Response, error) {
	url := "https://api.openai.com/v1/chat/completions"  // POST 请求的目标 URL

	if len(lib.Global.ApiProxyHost) > 0 {
		url = fmt.Sprintf("%s/v1/chat/completions", lib.Global.ApiProxyHost)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(conv.ToParams()))
	req.Header.Set("Content-Type", "application/json")                // 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apikey)) // 设置请求头

	client := &http.Client{Timeout: 2 * time.Minute}
	logger.Info(fmt.Sprintf("HTTP Request Body: %+v", string(conv.ToParams())))
	resp, err := client.Do(req) // 发送请求
	if err != nil {
		logger.Warning(fmt.Sprintf("Error sending HTTP request: %+v", err))
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	logger.Info(fmt.Sprintf("HTTP Response Status: %+v", resp.Status))

	// 读取响应体
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	logger.Info(fmt.Sprintf("HTTP Response Body:: %+v", buf.String()))

	var rsp Response
	err = json.Unmarshal(buf.Bytes(), &rsp)
	rsp.Status = resp.Status

	return &rsp, err
}
