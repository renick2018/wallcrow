package main

import (
	"github.com/gin-gonic/gin"
	"wallcrow/app"
	"wallcrow/lib"
	"wallcrow/lib/logger"
)

func main() {
	logger.Init(gin.Mode())

	//gin.SetMode(gin.DebugMode)
	lib.Init(gin.Mode())

	// 加载多个APP的路由配置
	app.Include(app.Routers)
	// 初始化路由
	r := app.Init()
	if err := r.Run(":8088"); err != nil {
		logger.Error("start web listener err: ", err.Error())
	}
}
