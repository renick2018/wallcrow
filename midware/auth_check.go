package midware

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"wallcrow/lib"
	"wallcrow/lib/logger"
)

var whiteList []string
var apiList []string

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, v := range whiteList {
			if strings.Index(path, fmt.Sprintf("/%s/", v)) == 0 {
				c.Next()
				return
			}
		}

		for _, v := range apiList {
			if strings.Index(path, fmt.Sprintf("/%s/", v)) == 0 {
				if checkApiParams(c) {
					c.Next()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "sign error"})
					c.Abort()
				}
				return
			}
		}

		if checkToken(c) {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "sign error"})
			c.Abort()
		}
	}
}

func checkToken(c *gin.Context) bool {
	token := c.Request.Header.Get("Authorization")
	// todo get cache
	var localToken = ""
	return token == localToken
}

func checkApiParams(c *gin.Context) bool {
	var params map[string]interface{}
	if c.Request.Method != "POST" {
		logger.Warning("auth parse err: api only support post request")
		return false
	}

	switch c.Request.Method {
	case "POST", "PUT":
		params = row2map(c)
	case "GET", "DELETE":
		params = query2map(c)
	}

	if params != nil || params["sign"] == nil || params["timestamp"] == nil{
		logger.Info(fmt.Sprintf("auth parse false"))
		return false
	}

	if reflect.TypeOf(params["sign"]).String() != "string"{
		return false
	}

	var timestamp int
	switch params["timestamp"].(type) {
	case float64, float32:
		timestamp = int(params["timestamp"].(float64))
	case string:
		var e error
		timestamp, e = strconv.Atoi(params["timestamp"].(string))
		if e != nil {
			return false
		}
	default:
		logger.Warning(fmt.Sprintf("auth parse error, timestamp type error"))
		return false
	}

	if time.Now().UnixMilli() < int64(timestamp) || time.Now().UnixMilli() - int64(timestamp) > lib.Global.ApiExpire {
		logger.Warning(fmt.Sprintf("auth parse error, timestamp expired or invalid"))
		return false
	}

	var sign = params["sign"].(string)
	params["sign"] = buildSalt(timestamp)

	bs, _ := json.Marshal(params)

	// 计算字符串的 MD5 值
	hash := md5.Sum(bs)

	// 将二进制 MD5 值转换为十六进制字符串
	token := hex.EncodeToString(hash[:])

	logger.Info(fmt.Sprintf("auth sign: %s, token:%s", sign, token))

	return token == sign
}

func buildSalt(timestamp int) string {
	var last = timestamp % 10
	var sb = strings.Builder{}
	for i, n := range strconv.Itoa(timestamp){
		if i < 3 {
			continue
		}
		if (last + int(n)) % 2 == 0 {
			sb.WriteString(fmt.Sprintf("%1f", math.Abs(float64(last - int(n)))))
		}
	}

	for i, n := range strconv.Itoa(timestamp){
		if i < 3 {
			continue
		}
		if (last + int(n)) % 2 != 0 {
			sb.WriteString(strconv.Itoa((last + int(n)) % 10))
		}
	}

	var saltStr = fmt.Sprintf("%s%s", lib.Global.ApiSalt, sb.String())

	logger.Info(fmt.Sprintf("timestamp: %d, new: %s, salt: %s", timestamp, sb, saltStr))

	// 计算字符串的 MD5 值
	hash := md5.Sum([]byte(saltStr))

	// 将二进制 MD5 值转换为十六进制字符串
	return hex.EncodeToString(hash[:])
}

func query2map(c *gin.Context) map[string]interface{} {
	var params = make(map[string]interface{})
	for _, item := range c.Params{
		params[item.Key] = item.Value
	}
	return params
}

func row2map(c *gin.Context) map[string]interface{} {
	var params = make(map[string]interface{})
	err := c.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		logger.Info(fmt.Sprintf("auth parse post params err: %+v", err))
		return nil
	}
	return params
}