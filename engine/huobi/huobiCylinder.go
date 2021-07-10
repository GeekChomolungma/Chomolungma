package huobi

import (
	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
)

type HuoBiCylinder struct {
}

func (HBCylinder *HuoBiCylinder) Ignite() {
	client := new(clients.AccountClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	resp, err := client.GetAccountInfo()
	if err != nil {
		applogger.Error("Get account error: %s", err)
	} else {
		applogger.Info("Get account, count=%d", len(resp))
		for _, result := range resp {
			applogger.Info("account: %+v", result)
		}
	}
}
