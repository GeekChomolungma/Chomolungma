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
func PlaceOrder(symbol, model, amount, price, source string) {
	client := new(clients.OrderClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	request := order.PlaceOrderRequest{
		AccountId: "3667382",
		Symbol:    symbol, //"btcusdt"
		Type:      model,  //"buy-limit"
		Source:    source, //"spot-api"
	}

	applogger.Info("PlaceOrder: Order model is %s", model)
	s, _ := db.CreateMarketDBSession()
	collectionName := fmt.Sprintf("%s-%s", "HB", "symbols")
	DBclient := s.DB("marketinfo").C(collectionName)
	sameSym := &common.SymbolFloat{}
	err := DBclient.Find(bson.M{"symbol": symbol}).One(sameSym)
	if err != nil {
		applogger.Error("PlaceOrder: Error, there is no symbol: %s could be traded! Please check HuoBi's available symbols", symbol)
		return
	}

	switch model {
	case "buy-market":
		// get precision of quote-currency
		// like usdt scale 8
		amountSeperates := strings.Split(amount, ".")
		request.Amount = amountSeperates[0]

	case "sell-market":
		// get precision of base-currency
		// like btc scale 6
		amountSeperates := strings.Split(amount, ".")
		rawDecimal := amountSeperates[1]

		baseCurrencyPrecision := sameSym.AmountPrecision
		applogger.Info("PlaceOrder: Symbol: %s, the amount precision is %d", symbol, baseCurrencyPrecision)
		if len(rawDecimal) < baseCurrencyPrecision {
			// amount too short
			request.Amount = amount
		} else {
			// cut the amount tail
			Decimal := rawDecimal[0:baseCurrencyPrecision]
			amount := fmt.Sprintf("%s.%s", amountSeperates[0], Decimal)
			request.Amount = amount
		}
		applogger.Info("btc amount to sell is %s", request.Amount)

	case "buy-limit":
		request.Price = price
		request.Amount = amount

	case "sell-limit":
		request.Price = price
		request.Amount = amount

	case "buy-limit-maker":
	case "sell-limit-maker":
	default:
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

func GetTimestamp() (int, error) {
	httpClient := new(clients.CommonClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.ApiServerHost,
	)
	timeMillsecs, err := httpClient.GetTimestamp()
	timestamp := timeMillsecs / 1000
	if err != nil {
		applogger.Error("Get GetTimestamp error: %s", err.Error())
		return 0, err
	}

	applogger.Info("GetTimestamp in second: %d", timestamp)
	return timestamp, nil
}
