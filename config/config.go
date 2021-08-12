package config

import (
	"fmt"
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
	SecretKey     []string
}

var HuoBiApiSetting = &HuoBiApiConf{}

type GatewayConf struct {
	GatewayHost    string
	GatewayTcpHost string
}

var GatewaySetting = &GatewayConf{}

type MarketSubConf struct {
	Symbols []string
	Periods []string
}

var MarketSubSetting = &MarketSubConf{}
var HBMarketSubList []string
var AccountMap = make(map[string]string)
var SecretMap = make(map[string]string)

// Setup
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
	mapTo(cfg, "marketsub", MarketSubSetting)

	// AccessKey
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

	// SecretKey
	if len(HuoBiApiSetting.SecretKey) != len(HuoBiApiSetting.AccountId) {
		applogger.Error("SecretKey not match AccountId, please check config.")
		panic("")
	}

	for i := 0; i < len(HuoBiApiSetting.SecretKey); i++ {
		if _, exist := SecretMap[HuoBiApiSetting.AccountId[i]]; !exist {
			// Not exist, push ak into accountmap
			SecretMap[HuoBiApiSetting.AccountId[i]] = HuoBiApiSetting.SecretKey[i]
		} else {
			applogger.Error("AccountId duplicated, please check config.")
			panic("")
		}
	}

	for _, syb := range MarketSubSetting.Symbols {
		for _, period := range MarketSubSetting.Periods {
			label := fmt.Sprintf("HB-%s-%s", syb, period)
			HBMarketSubList = append(HBMarketSubList, label)
		}
	}
	applogger.Info("HB marketinfo sub list length is %d", len(HBMarketSubList))
	applogger.Info("API and Secret Keys Loaded, there are %d accountkeys and %d secretKey.", len(HuoBiApiSetting.AccessKey), len(HuoBiApiSetting.SecretKey))
	applogger.Info("Config Setup Success.")
}

func mapTo(cfg *ini.File, section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}

func ReloadKeys() bool {
	applogger.Info("ReloadKeys...")
	cfg, err := ini.Load("./my.ini")
	if err != nil {
		applogger.Error("Fail to parse '../my.ini': %v", err)
	}

	mapTo(cfg, "huobi", HuoBiApiSetting)

	// AccessKey
	if len(HuoBiApiSetting.AccessKey) != len(HuoBiApiSetting.AccountId) {
		applogger.Error("AccessKey not match AccountId, please check config.")
		return false
	}

	for i := 0; i < len(HuoBiApiSetting.AccessKey); i++ {
		if _, exist := AccountMap[HuoBiApiSetting.AccountId[i]]; !exist {
			// Not exist, push ak into accountmap
			AccountMap[HuoBiApiSetting.AccountId[i]] = HuoBiApiSetting.AccessKey[i]
		}
	}

	// SecretKey
	if len(HuoBiApiSetting.SecretKey) != len(HuoBiApiSetting.AccountId) {
		applogger.Error("SecretKey not match AccountId, please check config.")
		return false
	}

	for i := 0; i < len(HuoBiApiSetting.SecretKey); i++ {
		if _, exist := SecretMap[HuoBiApiSetting.AccountId[i]]; !exist {
			// Not exist, push ak into accountmap
			SecretMap[HuoBiApiSetting.AccountId[i]] = HuoBiApiSetting.SecretKey[i]
		}
	}
	applogger.Info("API and Secret Key map: %v,%v.", AccountMap, SecretMap)
	applogger.Info("API and Secret Keys reloaded, there are %d accountkeys and %d secretKey.", len(HuoBiApiSetting.AccessKey), len(HuoBiApiSetting.SecretKey))
	return true
}
