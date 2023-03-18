package model

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

type Conversation struct {
	BaseModel         `json:",inline"`
	UniqueID          string    `json:"unique_id" gorm:"index,unique,comment:唯一id"`
	Nickname          string    `json:"nickname" gorm:"comment:昵称"`
	Model             string    `json:"model" gorm:"comment:模型"`
	Temperature       float32   `json:"temperature" gorm:"comment:混乱/随机0～2"`
	MaxResponseTokens int       `json:"max_response_tokens" gorm:"comment:最大回复长度，max: MaxTokens - prompt"` // 4096
	MaxTokens         int       `json:"max_tokens" gorm:"comment:上下文长度限制 max: 4096"`
	N                 int       `json:"n" gorm:"comment: 生成个数"`
	Messages          []Message `json:"messages"  gorm:"-"`
	SystemPrompt      string    `json:"system_prompt" gorm:"comment:系统提示词"`
	//ContextSwitch     bool      `json:"context_switch"  gorm:"comment:上下文开关"`
	Tokens int `json:"-" gorm:"-"`
}

type Role string

const (
	User      Role = "user"      // 用户
	System    Role = "system"    // 设定
	Assistant Role = "assistant" // GPT
)

type Message struct {
	BaseModel `json:",inline"`
	ConvID    uint `json:"conv_id,omitempty" gorm:""`
	//ParentID     uint   `json:"parent_id,omitempty" gorm:"comment: 父消息id"`
	Role         Role   `json:"role" gorm:"comment:角色"`
	Content      string `json:"content" gorm:"comment:内容"`
	Tokens       int    `json:"tokens,omitempty" gorm:"comment:消耗字数"`
	FinishReason string `json:"finish_reason,omitempty" gorm:"comment:gpt停止原因"`
}

type PicturePrompt struct {
	BaseModel `json:",inline"`
	Prompt    string `json:"prompt" gorm:"comment: 提示词"`
	Urls      string `json:"url" gorm:"comment: 图片地址"`
	SourceUrl string `json:"source_url" gorm:"comment:原图"`
	Size      string `json:"size" gorm:"comment:分辨率"`
	Number    int    `json:"number" gorm:"comment:数量"`
}

func (c *Conversation) ToParams() []byte {
	var messages = make([]Message, 0)

	if len(c.SystemPrompt) == 0 {
		c.SystemPrompt = fmt.Sprintf("You are ChatGPT, a large language model trained by OpenAI. Answer as concisely as possible.\nKnowledge cutoff: 2021-09-01\nCurrent date: " + time.Now().Format("2006-01-02"))
	}

	messages = append(messages, Message{Role: System, Content: c.SystemPrompt})
	for _, item := range c.Messages {
		messages = append(messages, Message{Role: item.Role, Content: item.Content})
	}
	data := make(map[string]interface{})
	//data["n"] = c.N
	data["model"] = c.Model
	data["temperature"] = c.Temperature
	data["max_tokens"] = int(math.Min(float64(c.MaxResponseTokens), float64(c.MaxTokens-c.Tokens)))
	data["top_p"] = 1
	data["presence_penalty"] = 1
	data["stream"] = false
	data["messages"] = messages
	bs, _ := json.Marshal(data) // POST 请求的数据
	return bs
}
