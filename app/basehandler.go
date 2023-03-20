package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"wallcrow/lib/logger"
)

func response(ctx *gin.Context, code int, message string, data ...interface{})  {
	if code != 0 || len(message) > 0 {
		logger.Debug(fmt.Sprintf("response %d %s", code, message))
	}
	var rsp = gin.H{}
	rsp["code"] = code
	rsp["message"] = message
	if len(data) > 0 {
		rsp["response"] = data[0]
	}
	ctx.JSON(http.StatusOK, rsp)
}