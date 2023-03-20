package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"wallcrow/component/chatgpt"
	"wallcrow/lib"
	"wallcrow/model"
)

type chatroom struct {
}

func (chatroom) saveOrUpdate(ctx *gin.Context) {

}

func (chatroom) show(ctx *gin.Context) {

}

func (chatroom) list(ctx *gin.Context) {

}

func (chatroom) clear(ctx *gin.Context) {

}

func (chatroom) ask(ctx *gin.Context) {
	var params struct {
		UniqueID string                 `json:"conversationId"`
		Message  string                 `json:"message"`
		Options  map[string]interface{} `json:"options"`
	}
	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil || len(params.UniqueID) == 0 {
		response(ctx, -1, "params error")
		return
	}

	var conv model.Conversation

	lib.DB.Model(model.Conversation{}).Where("unique_id = ?", params.UniqueID).Find(&conv)

	if conv.ID == 0 {
		conv = model.Conversation{
			UniqueID:          params.UniqueID,
			Model:             "gpt-3.5-turbo",
			Temperature:       0.8,
			MaxResponseTokens: 1000,
			MaxTokens:         2048,
			N:                 1,
			Messages:          make([]model.Message, 0),
		}
		lib.DB.Model(model.Conversation{}).Save(&conv)
	}

	rsp, err := chatgpt.Ask(&conv, params.Message)
	if err != nil || rsp.Error != nil {
		response(ctx, -1, fmt.Sprintf("ask openai error: %+v, detail: %+v", err, rsp.Error))
		return
	}

	var data = make(map[string]interface{})
	data["conversationId"] = params.UniqueID
	data["response"] = rsp.Choices[0].Message.Content

	response(ctx, 0, "", data)
}
