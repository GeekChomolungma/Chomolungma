package huobi

import (
	"fmt"
	"strconv"
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
		Symbol:    symbol, // "btcusdt"
		Type:      model,  // "buy-limit"
		Source:    source, // "spot-api"
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

	// precision setting
	baseCurLimitPrecision := sameSym.AmountPrecision // for limit
	qutoCurLimitPrecision := sameSym.PricePrecision  // for limit
	amountMarketPrecision := sameSym.AmountPrecision // for market sell
	valueMarketPrecision := sameSym.ValuePrecision   // for market buy
	applogger.Info("PlaceOrder: Symbol: %s, LIMIT Model: [amount precision %d, price precision: %d]. MARKET Model: [amount precision %d, value precision: %d]. MARKET Model",
		symbol, baseCurLimitPrecision, qutoCurLimitPrecision, amountMarketPrecision, valueMarketPrecision)

	// min and max value setting
	limitOrderMinOrderAmt := sameSym.LimitOrderMinOrderAmt
	limitOrderMaxOrderAmt := sameSym.LimitOrderMaxOrderAmt
	sellMarketMinOrderAmt := sameSym.SellMarketMinOrderAmt
	sellMarketMaxOrderAmt := sameSym.SellMarketMaxOrderAmt
	buyMarketMaxOrderValue := sameSym.BuyMarketMaxOrderValue
	minOrderValue := sameSym.MinOrderValue
	maxOrderValue := sameSym.MaxOrderValue
	applogger.Info("PlaceOrder: Symbol: %s, limitOrderMinOrderAmt:%f, limitOrderMaxOrderAmt:%f, sellMarketMinOrderAmt: %f, sellMarketMaxOrderAmt: %f, buyMarketMaxOrderValue:%f, minOrderValue:%f, maxOrderValue:%f",
		symbol,
		limitOrderMinOrderAmt, limitOrderMaxOrderAmt, sellMarketMinOrderAmt, sellMarketMaxOrderAmt,
		buyMarketMaxOrderValue,
		minOrderValue, maxOrderValue)

	switch model {
	case "buy-market":
		// get precision of quote-currency
		// like usdt scale 8
		// check min amount for buy
		passed, err := checkMinAndMaxValAmt(amount, minOrderValue, buyMarketMaxOrderValue)
		if err != nil || !passed {
			applogger.Error("buy-market: check min or max not passed.")
			return
		}
		// check amount precision
		Amt, amtErr := checkAmtPrecision(amount, valueMarketPrecision)
		if amtErr != nil {
			applogger.Error("buy-market: checkAmtPrecision error: %s", amtErr.Error())
			return
		}
		request.Amount = Amt
		applogger.Info("buy-market, usdt amount used to buy is %s", request.Amount)

	case "sell-market":
		// get precision of base-currency
		// like btc scale 6
		// check min amount for sell
		passed, err := checkMinAndMaxValAmt(amount, sellMarketMinOrderAmt, sellMarketMaxOrderAmt)
		if err != nil || !passed {
			applogger.Error("sell-market: check min or max not passed.")
			return
		}
		// check amount precision
		Amt, amtErr := checkAmtPrecision(amount, amountMarketPrecision)
		if amtErr != nil {
			applogger.Error("sell-market: checkAmtPrecision error: %s", amtErr.Error())
			return
		}
		request.Amount = Amt
		applogger.Info("sell-market, btc amount to sell is %s", request.Amount)

	case "buy-limit", "sell-limit":
		// fistly, check amount
		// check min amount
		passed, err := checkMinAndMaxValAmt(amount, limitOrderMinOrderAmt, limitOrderMaxOrderAmt)
		if err != nil || !passed {
			applogger.Error("limit: check amount min or max not passed.")
			return
		}
		// check amount precision
		Amt, amtErr := checkAmtPrecision(amount, baseCurLimitPrecision)
		if amtErr != nil {
			applogger.Error("limit: baseCurLimitPrecision error: %s", amtErr.Error())
			return
		}
		request.Amount = Amt
		applogger.Info("limit, amount is %s", request.Amount)

		// then, check price
		Price, priceErr := checkPricePrecision(price, qutoCurLimitPrecision)
		if priceErr != nil {
			applogger.Error("limit: qutoCurLimitPrecision error: %s", amtErr.Error())
			return
		}
		request.Price = Price
		applogger.Info("limit, price is %s", request.Price)

		// finally, check production of price * amount
		passed = checkProduction(request.Amount, request.Price, minOrderValue, buyMarketMaxOrderValue)
		if !passed {
			applogger.Error("limit: check production not passed, min value is %f, max value is %f", minOrderValue, buyMarketMaxOrderValue)
			return
		}

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

func checkPricePrecision(price string, precision int) (string, error) {
	var Price string
	priceSeperates := strings.Split(price, ".")
	rawDecimal := priceSeperates[1]
	if len(rawDecimal) >= precision {
		// cut the price tail
		Decimal := rawDecimal[0:precision]
		price = fmt.Sprintf("%s.%s", priceSeperates[0], Decimal)
	}
	Price = price
	return Price, nil
}

func checkAmtPrecision(amount string, precision int) (string, error) {
	var Amount string
	amountSeperates := strings.Split(amount, ".")
	rawDecimal := amountSeperates[1]
	if len(rawDecimal) >= precision {
		// cut the amount tail
		Decimal := rawDecimal[0:precision]
		amount = fmt.Sprintf("%s.%s", amountSeperates[0], Decimal)
	}
	Amount = amount
	return Amount, nil
}

func checkMinAndMaxValAmt(amount string, minV, maxV float64) (bool, error) {
	amountF, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return false, err
	}
	if amountF < minV || amountF > maxV {
		// amount over limit.
		return false, nil
	}
	return true, nil
}

func checkProduction(amount, price string, minV, maxV float64) bool {
	amountF, _ := strconv.ParseFloat(amount, 64)
	priceF, _ := strconv.ParseFloat(price, 64)
	product := amountF * priceF
	if product < minV || product > maxV {
		// value over limit.
		return false
	}
	return true
}

func CancelOrderById(orderID string) {
	client := new(clients.OrderClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	resp, err := client.CancelOrderById(orderID)
	if err != nil {
		applogger.Error(err.Error())
	} else {
		switch resp.Status {
		case "ok":
			applogger.Info("Cancel order successfully, order id: %s", resp.Data)
		case "error":
			applogger.Info("Cancel order error: %s", resp.ErrorMessage)
		}
	}
}

func GetOrderById(orderID string) {
	client := new(clients.OrderClient).Init(
		config.GatewaySetting.GatewayHost,
		config.HuoBiApiSetting.AccessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)
	resp, err := client.GetOrderById(orderID)
	if err != nil {
		applogger.Error(err.Error())
	} else {
		switch resp.Status {
		case "ok":
			applogger.Info("Query order successfully, order info: %v", resp.Data)
		case "error":
			applogger.Info("Query order error: %s", resp.ErrorMessage)
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
		applogger.Error("Ignite: Initially Get GetTimestamp error: %s", err.Error())
		return 0, err
	}

	applogger.Info("Ignite: Initially GetTimestamp in second: %d", timestamp)
	return timestamp, nil
}
