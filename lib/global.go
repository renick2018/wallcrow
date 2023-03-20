package lib

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"wallcrow/lib/logger"
)

var Path string
var DB *gorm.DB
var Rds *redis.Client

var Global struct {
	ApiProxyHost string   `yaml:"api_proxy_host"`
	ApiProxyHostTs string `yaml:"api_proxy_host_ts"`
	ApiSalt      string   `yaml:"api_salt"`
	ApiExpire    int64    `yaml:"api_expire"`
	Emails       []string `yaml:"emails"` // alert emails
	EmailServer  struct {
		Sender   string `yaml:"sender"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
	} `yaml:"email_server"`

	MysqlConf struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
	} `yaml:"mysql_conf"`

	RedisConf struct {
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
	} `yaml:"redis_conf"`
}

func Init(env string) {
	Path, _ = os.Getwd()
	var confName = "debug.yml"
	switch env {
	case gin.TestMode:
		confName = "test.yml"
	case gin.ReleaseMode:
		confName = "release.yml"
	}
	fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", Path, confName))
	if err != nil {
		logger.Error(fmt.Sprintf("load conf file error: %+v", err))
		return
	}

	if e := yaml.Unmarshal(fileData, &Global); e != nil {
		logger.Error(fmt.Sprintf("unmarshal conf file error: %+v", err))
	}
	initDB()
	initRedis()
	logger.Info(fmt.Sprintf("load %s conf over", gin.Mode()))
}

func initDB() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/wallcrow?charset=utf8mb4&parseTime=True&loc=Local",
		Global.MysqlConf.User,
		Global.MysqlConf.Password,
		Global.MysqlConf.Host,
		Global.MysqlConf.Port,
	)
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logger.Warning(fmt.Sprintf("init database error: %v, conf: %+v", err, Global.MysqlConf))
	}
}

func initRedis() {
	Rds = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", Global.RedisConf.Host, Global.RedisConf.Port),
		Password: Global.RedisConf.Password,
		DB:       0,
	})
	if _, err := Rds.Ping().Result(); err != nil {
		logger.Warning(fmt.Sprintf("init redis error: %v, conf: %+v", err, Global.RedisConf))
	}
}
