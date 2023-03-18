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
	ctx.JSON(http.StatusOK, gin.H{
		"code": code,
		"message": message,
		"data": data,
	})
}