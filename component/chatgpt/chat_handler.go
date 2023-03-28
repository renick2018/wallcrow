package chatgpt

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"wallcrow/lib"
	"wallcrow/lib/email"
	"wallcrow/lib/logger"
	"wallcrow/model"
)

var convMap = convLock{
	convMap:   make(map[string]*sync.Mutex),
	locker:    sync.Mutex{},
	apiLocker: sync.Mutex{},
}

type convLock struct {
	apikey    *string
	convMap   map[string]*sync.Mutex
	locker    sync.Mutex
	apiLocker sync.Mutex
}

func (c *convLock) item(uid string) *sync.Mutex {
	if _, ex := c.convMap[uid]; ex {
		return c.convMap[uid]
	}
	c.locker.Lock()
	defer c.locker.Unlock()

	c.convMap[uid] = &sync.Mutex{}
	return c.convMap[uid]
}

func (c *convLock) lock(convUid string) {
	c.item(convUid).Lock()
}

func (c *convLock) unlock(convUid string) {
	c.convMap[convUid].Unlock()
}

func (c *convLock) fetchApikey() *string {
	if c.apikey == nil {
		c.apiLocker.Lock()
		defer c.apiLocker.Unlock()
		if c.apikey == nil {
			lib.DB.Model(model.OpenaiAccount{}).Select("apikey").Where("use_up = false").Limit(1).Scan(&c.apikey)
		}
	}

	return c.apikey
}

func Ask(conv *model.Conversation, message string) (*Response, error) {
	convMap.lock(conv.UniqueID)
	defer func() {
		convMap.unlock(conv.UniqueID)
	}()

	// fetch apikey
	var apikey = convMap.fetchApikey()

	if apikey == nil {
		return nil, errors.New("no available apikey")
	}

	var rsp *Response
	var err error
	if len(lib.Global.ApiProxyHostTs) > 0 {
		rsp, err = askByTs(message, conv.UniqueID, *apikey)
	}else {
		rsp, err = askByProxy(conv, message, *apikey)
	}

	if err != nil || rsp == nil || rsp.Error != nil {
		if rsp == nil {
			logger.Warning(fmt.Sprintf("openai apikey %s request openai failed, error: %+v", *apikey, err))
		} else {
			logger.Warning(fmt.Sprintf("openai apikey %s request openai failed, error: %+v", *apikey, rsp.Error))
			if rsp.Error.Type == "insufficient_quota" || (rsp.Error.Code != nil && *rsp.Error.Code == "account_deactivated"){
				convMap.apikey = nil
				var status = rsp.Error.Type
				if rsp.Error.Code != nil {
					status = *rsp.Error.Code
				}
				lib.DB.Model(model.OpenaiAccount{}).Where("apikey = ?", *apikey).Updates(map[string]interface{}{"use_up": true, "status": status})
				go email.Alert(fmt.Sprintf("wallcrow %s", rsp.Error.Type), fmt.Sprintf("openai apikey <strong>%s</strong> is used up<br> error type: <strong>%s</strong><br> message: %s", *apikey, rsp.Error.Type, rsp.Error.Message))
				apikey = convMap.fetchApikey()
				if apikey == nil {
					return nil, errors.New("api quota insufficient")
				}
			}
		}
		time.Sleep(5 * time.Second)
		if len(lib.Global.ApiProxyHostTs) > 0 {
			rsp, err = askByTs(message, conv.UniqueID, *apikey)
		}else {
			rsp, err = askByProxy(conv, message, *apikey)
		}
	}

	return rsp, err
}

func askByTs(message, convUid, apikey string) (*Response, error) {
	messageIdMap := "hash:conv:messageid"
	cmd := lib.Rds.HGet(messageIdMap, convUid)
	messageId := cmd.Val()
	rsp, err := askTs(message, messageId, apikey)
	if err != nil || rsp == nil || rsp.Error != nil {
		return rsp, err
	}

	lib.Rds.HSet(messageIdMap, convUid, rsp.MessageId)

	// todo save message logs

	return rsp, err
}

func askByProxy(conv *model.Conversation, message, apikey string) (*Response, error) {
	// fetch conv context messages from redis
	var convRdsKey = ":chatcontext"
	var messages []model.Message
	if cmd := lib.Rds.HGet(convRdsKey, conv.UniqueID); cmd != nil {
		err := json.Unmarshal([]byte(cmd.Val()), &messages)
		if err != nil{
			logger.Warning(fmt.Sprintf("fetch message error: %+v\n%s", err, cmd.Val()))
		}
	}

	logger.Info(fmt.Sprintf("find messages: %d\n%+v", len(messages), messages))

	var tokens = 0
	for i := len(messages) - 1; i >= 0; i-- {
		var item = messages[i]
		//logger.Debug(fmt.Sprintf("token %d, cut: %d, index: %d, messages: %d, %s", tokens, conv.MaxTokens-conv.MaxResponseTokens, i, item.Tokens, item.Content))
		tokens += item.Tokens
		if tokens >= conv.MaxTokens-conv.MaxResponseTokens {
			tokens -= item.Tokens
			break
		}
		conv.Messages = append([]model.Message{item}, conv.Messages...)
	}

	if len(conv.Messages) > 0 && conv.Messages[0].Role == model.Assistant {
		conv.Messages = conv.Messages[1:]
	}

	conv.Tokens = tokens

	logger.Debug(fmt.Sprintf("token: %d, messages: %d", conv.Tokens, len(conv.Messages)))

	var question = model.Message{Role: model.User, Content: message}
	conv.Messages = append(conv.Messages, question)

	logger.Info(fmt.Sprintf("%s %s", apikey, message))

	rsp, err := ask(conv, apikey)

	logger.Warning(fmt.Sprintf("openai apikey request res: %+v", rsp))

	if err != nil || rsp == nil || len(rsp.Choices) == 0 {
		return rsp, err
	}

	question.Tokens = rsp.Usage.PromptTokens - conv.Tokens
	lib.DB.Model(model.Message{}).Save(&question)
	conv.Messages[len(conv.Messages)-1] = question
	for _, item := range rsp.Choices {
		var msg = model.Message{
			ConvID: conv.ID,
			Role:         item.Message.Role,
			Content:      item.Message.Content,
			Tokens:       rsp.Usage.CompletionTokens / len(rsp.Choices),
			FinishReason: item.FinishReason,
		}
		conv.Messages = append(conv.Messages, msg)
		lib.DB.Model(model.Message{}).Save(&msg)
	}

	//todo test is work or not
	bs, _ := json.Marshal(conv.Messages)
	lib.Rds.HSet(convRdsKey, conv.UniqueID, string(bs))
	logger.Info(fmt.Sprintf("【%s】", conv.UniqueID))
	for _, item := range conv.Messages {
		logger.Info(fmt.Sprintf("【%10s】%03d: %s", item.Role, item.Tokens, item.Content))
	}
	return rsp, err
}