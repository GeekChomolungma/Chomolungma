package config

import (
	"log"

	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"github.com/go-ini/ini"
)

type MongoConf struct {
	MarketUrl  string
	AccountUrl string
	RootUrl    string
}

var MongoSetting = &MongoConf{}

type Server struct {
	Host string
}

var ServerSetting = &Server{}

type HuoBiApiConf struct {
	ApiServerHost string
	AccessKey     []string
	AccountId     []string
	SubUid        int
	SubUids       string
	SecretKey     string
}

var HuoBiApiSetting = &HuoBiApiConf{}

type GatewayConf struct {
	GatewayHost    string
	GatewayTcpHost string
}

var GatewaySetting = &GatewayConf{}

var AccountMap = make(map[string]string)

// Setup 启动配置
func Setup() {
	applogger.Info("Config Loading...")
	cfg, err := ini.Load("./my.ini")
	if err != nil {
		applogger.Error("Fail to parse '../my.ini': %v", err)
	}

	mapTo(cfg, "mongo", MongoSetting)
	mapTo(cfg, "server", ServerSetting)
	mapTo(cfg, "gateway", GatewaySetting)
	mapTo(cfg, "huobi", HuoBiApiSetting)
	if len(HuoBiApiSetting.AccessKey) != len(HuoBiApiSetting.AccountId) {
		applogger.Error("AccessKey not match AccountId, please check config.")
		panic("")
	}

	for i := 0; i < len(HuoBiApiSetting.AccessKey); i++ {
		if _, exist := AccountMap[HuoBiApiSetting.AccountId[i]]; !exist {
			// Not exist, push ak into accountmap
			AccountMap[HuoBiApiSetting.AccountId[i]] = HuoBiApiSetting.AccessKey[i]
		} else {
			applogger.Error("AccountId duplicated, please check config.")
			panic("")
		}
	}
	applogger.Info("AccountMap Loaded, there are %d accountkeys in.", len(HuoBiApiSetting.AccessKey))
	applogger.Info("Config Setup Success.")
}

func mapTo(cfg *ini.File, section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}
