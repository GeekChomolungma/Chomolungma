package config

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
