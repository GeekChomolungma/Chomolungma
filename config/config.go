package config

import (
	"log"

	"github.com/go-ini/ini"
)

type MySQLConf struct {
	Host     string
	User     string
	PassWord string
	DataBase string
}

var MySQLSetting = &MySQLConf{}

type Server struct {
	Host string
}

var ServerSetting = &Server{}

type HuoBiApiConf struct {
	ApiServerHost string
	AccessKey     string
	AccountId     string
	SubUid        int
	SubUids       string
	SecretKey     string
}

var HuoBiApiSetting = &HuoBiApiConf{}

type GatewayConf struct {
	GatewayHost string
}

var GatewaySetting = &GatewayConf{}

// Setup 启动配置
func Setup() {
	cfg, err := ini.Load("./my.ini")
	if err != nil {
		log.Fatalf("Fail to parse '../my.ini': %v", err)
	}

	mapTo(cfg, "mysql", MySQLSetting)
	mapTo(cfg, "server", ServerSetting)
	mapTo(cfg, "gateway", GatewaySetting)
	mapTo(cfg, "huobi", HuoBiApiSetting)
}

func mapTo(cfg *ini.File, section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}
