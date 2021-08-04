package huobi

import (
	"fmt"
	"strings"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/db"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/account"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/common"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/order"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2/bson"
)

// -------------------------------------------------------------ACCOUNT-------------------------------------------------------
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
		applogger.Error("Get account error: %s", err.Error())
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
		applogger.Error("Cannot get account balance: %s", err.Error())
		return nil, err
	} else {
		applogger.Info("Got account balance")
		return resp, nil
	}
}

// -------------------------------------------------------------ORDER-------------------------------------------------------
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
	applogger.Info("model is %s", model)
	if model == "buy-market" {
		// usdt scale 8
		request.Amount = amountSeperates[0]
	} else {
		if model == "sell-market" {
			// btc scale 6
			rawDecimal := amountSeperates[1]
			Decimal := rawDecimal[0:6]
			amount := fmt.Sprintf("%s.%s", amountSeperates[0], Decimal)
			request.Amount = amount
			applogger.Info("btc amount to sell is %s", amount)
		} else {
			request.Amount = amount
		}
	}

	if model == "buy-market" || model == "sell-market" {
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

// -------------------------------------------------------------COMMON-------------------------------------------------------
func GetSymbols() ([]common.Symbol, error) {
	httpClient := new(clients.CommonClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.ApiServerHost,
	)
	symbols, err := httpClient.GetSymbols()
	if err != nil {
		applogger.Error("Get GetSymbols error: %s", err.Error())
		return nil, err
	}

	applogger.Info("Get symbols, count=%d", len(symbols))
	return symbols, nil
}

func querySymbolsAndWriteDisk() {
	// connect market db
	s, err := db.CreateMarketDBSession()
	collectionName := fmt.Sprintf("%s-%s", "HB", "symbols")
	client := s.DB("marketinfo").C(collectionName)
	mgoSessionMap[collectionName] = s
	if err != nil {
		applogger.Error("Failed to connection db: %s", err.Error())
		return
	}

	symbols, err := GetSymbols()
	if err != nil {
		applogger.Error("querySymbolsAndWriteDisk: Get GetSymbols error: %s", err.Error())
		return
	}

	for _, s := range symbols {
		sameSym := &common.SymbolFloat{}
		err := client.Find(bson.M{"symbol": s.Symbol}).One(sameSym)
		if err != nil {
			// if not exist, insert
			sf := s.SymbolToFloat()
			err = client.Insert(sf)
			if err != nil {
				applogger.Error("querySymbolsAndWriteDisk: Failed to connection db: %s", err.Error())
			} else {
				applogger.Info("Success to write symbol %s into db", s.Symbol)
			}
		} else {
			// update it
			selector := bson.M{"symbol": s.Symbol}
			sf := s.SymbolToFloat()
			err := client.Update(selector, sf)
			if err != nil {
				applogger.Error("querySymbolsAndWriteDisk: Failed to update symbol: %s", err.Error())
			}
		}
	}
}
