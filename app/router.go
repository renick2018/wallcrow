package app

import (
	"github.com/gin-gonic/gin"
	"wallcrow/midware"
)

type Option func(engine *gin.Engine)

var options []Option

func Include(opts ...Option) {
	options = append(options, opts...)
}

// Init 初始化
func Init() *gin.Engine {
	r := gin.New()
	r.Use(midware.Cors())
	//r.Use(midware.Auth())
	for _, opt := range options {
		opt(r)
	}
	return r
}

func Routers(e *gin.Engine) {
	var openai = e.Group("/openai")
	{
		// account
		{
			openai.GET("/account", account{}.list) // 获取账号列表
			openai.PUT("/account", account{}.saveOrUpdate) // 添加账号
			openai.DELETE("/account", account{}.wipe) // 删除账号
		}

		// chatgpt
		var chat = openai.Group("/chat")
		{
			chat.GET("", chatroom{}.list) // 聊天室列表
			chat.GET("/:id", chatroom{}.show) // 聊天室信息+消息记录
			chat.PUT("", chatroom{}.saveOrUpdate) // 聊天室配置
			chat.POST("/ask", chatroom{}.ask) // 询问
			chat.DELETE("/:id", chatroom{}.clear) // 清理聊天记录
		}

		//dalle
		picture := openai.Group("/dalle")
		{
			picture.GET("", dalle{}.list) //查看图片
			picture.GET("/:id", dalle{}.show) //查看图片
			picture.POST("", dalle{}.paint) // 新建图片
		}
	}
}
