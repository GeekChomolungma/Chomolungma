package huobi

import (
	"strings"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/account"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/order"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
)

// GetAccountInfo return the account info
func GetAccountInfo() {
	httpClient := new(clients.AccountClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	resp, err := httpClient.GetAccountInfo()
	if err != nil {
		applogger.Error("Get account error: %s", err)
	} else {
		applogger.Info("Get account, count=%d", len(resp))
		for _, result := range resp {
			applogger.Info("account: %+v", result)
		}
	}
}

func GetAccountBalance() (*account.AccountBalance, error) {
	httpClient := new(clients.AccountClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	resp, err := httpClient.GetAccountBalance("3667382")
	if err != nil {
		applogger.Error("Cannot get account balance: %s", err)
		return nil, err
	} else {
		applogger.Info("Get account balance, %v", resp)
		return resp, nil
	}
}

func PlaceOrder(model, price, amount string) {
	client := new(clients.OrderClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	request := order.PlaceOrderRequest{
		AccountId: "3667382",
		Type:      model, //"buy-limit",
		Source:    "spot-api",
		Symbol:    "btcusdt",
	}
	amountSeperates := strings.Split(amount, ".")
	//applogger.Info("amount is %s", amountSeperates[0])
	request.Amount = amountSeperates[0]
	if model == "buy-market" {
		applogger.Info("market order, no price, req is: %v", request)
	} else {
		request.Price = price
	}

	resp, err := client.PlaceOrder(&request)
	if err != nil {
		applogger.Error(err.Error())
	} else {
		switch resp.Status {
		case "ok":
			applogger.Info("Place order successfully, order id: %s", resp.Data)
		case "error":
			applogger.Error("Place order error: %s", resp.ErrorMessage)
		}
	}
}
