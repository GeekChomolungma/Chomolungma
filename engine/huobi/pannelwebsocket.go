package huobi

import (
	"fmt"
	"sync"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/db"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/marketwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/orderwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/auth"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/market"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/order"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2/bson"
)

var PreviousSyncTimeMap sync.Map

type periodUnit string

const (
	Period_1min  periodUnit = "1min"
	Period_5min  periodUnit = "5min"
	Period_15min periodUnit = "15min"
	Period_30min periodUnit = "30min"
	Period_60min periodUnit = "60min"
	Period_4hour periodUnit = "4hour"
	Period_1day  periodUnit = "1day"
	Period_1week periodUnit = "1week"
	Period_1mon  periodUnit = "1mon"
	Period_1year periodUnit = "1year"
)

// -------------------------------------------------------------MARKET-------------------------------------------------------
func subscribeMarketInfo(symbol string, period periodUnit) {
	var previousTick *market.Tick

	// connect market db
	s, err := db.CreateMarketDBSession()
	collectionName := fmt.Sprintf("HB-%s-%s", symbol, string(period))
	client := s.DB("marketinfo").C(collectionName)
	mgoSessionMap[collectionName+"-subscribeMarketInfo"] = s
	if err != nil {
		applogger.Error("Failed to connection #%s db: %s", symbol, err.Error())
		return
	}

	// websocket
	wsClient := new(marketwebsocketclient.CandlestickWebSocketClient).Init(config.GatewaySetting.GatewayHost)

	wsClient.SetHandler(
		func() {
			wsClient.Subscribe(symbol, string(period), "2118")
		},
		func(response interface{}) {
			resp, ok := response.(market.SubscribeCandlestickResponse)
			if ok {
				if &resp != nil {
					if resp.Tick != nil {
						t := resp.Tick
						// applogger.Info("Candlestick update, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
						// 	t.Id, t.Count, t.Vol, t.Open, t.Close, t.Low, t.High)

						tickerRetrive := &market.TickFloat{}
						err := client.Find(bson.M{"id": t.Id}).One(tickerRetrive)
						if err != nil {
							// if not exist, insert
							// insert new data
							tickerWrite := t.TickToFloat()
							err = client.Insert(tickerWrite)
							if err != nil {
								applogger.Error("Failed to Insert #%s data : %s", symbol, err.Error())
							} else {
								applogger.Info("New      #%s Data  Pushed into DB: id: %d", symbol, t.Id)
							}

							if previousTick != nil {
								// update the previous data
								selector := bson.M{"id": previousTick.Id}
								previousTickFloat := previousTick.TickToFloat()
								err := client.Update(selector, previousTickFloat)
								if err != nil {
									applogger.Error("Failed to Update Previous #%s Data to db: %s", symbol, err.Error())
								} else {
									applogger.Info("Previous #%s Data Updated into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										symbol,
										previousTick.Id, previousTick.Count, previousTick.Vol,
										previousTick.Open, previousTick.Count, previousTick.Low, previousTick.High)
								}
							}
						}
						// update previous tick
						previousTick = t

						// add PreviousSyncTime into map
						PreviousSyncTimeMap.Store(collectionName, t.Id)
					}

					if resp.Data != nil {
						applogger.Info("WebSocket returned data, count=%d", len(resp.Data))
						for _, t := range resp.Data {
							applogger.Info("Candlestick data, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
								t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)

							tickerRetrive := &market.TickFloat{}
							err := client.Find(bson.M{"id": t.Id}).One(tickerRetrive)
							if err != nil {
								// if not exist, insert
								applogger.Error("not exist, insert")
								tickerWrite := t.TickToFloat()
								err = client.Insert(tickerWrite)
								if err != nil {
									applogger.Error("Failed to connection db: %s", err.Error())
								} else {
									applogger.Info("Candlestick data write to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
								}
							}
						}
					}
				}
			} else {
				applogger.Warn("Unknown response: %v", resp)
			}

		})

	wsClient.Connect(true)
	wsCandlestickClientMap[collectionName+"-subscribeMarketInfo"] = wsClient
}

// flowWindowMarketInfo returns tickers, and the max tickers number is 300 once.
// So, startTime and toTime should be constrain for:
//                   (toTime - startTime)/period < 300
func flowWindowMarketInfo(symbol string, period periodUnit, startTime int64, toTime int64) {
	// make a time window
	timeWindow, err := makeTimeWindow(period, startTime, toTime)
	if err != nil {
		return
	}
	applogger.Info("timeWindow length is: %d, data is: %d",
		len(timeWindow), timeWindow)
	// connect market db
	s, err := db.CreateMarketDBSession()
	collectionName := fmt.Sprintf("HB-%s-%s", symbol, string(period))
	client := s.DB("marketinfo").C(collectionName)
	mgoSessionMap[collectionName+"-flowWindowMarketInfo"] = s
	if err != nil {
		applogger.Error("Failed to connection db: %s", err.Error())
		return
	}

	// websocket
	wsClient := new(marketwebsocketclient.CandlestickWebSocketClient).Init(config.GatewaySetting.GatewayHost)
	wsClient.SetHandler(
		func() {
			// multi request for data returned under 300 once.
			// whatever the period is, request should make sure that
			// the res return data length less than 300.
			for _, timeEle := range timeWindow {
				wsClient.Request(symbol, string(period), timeEle[0], timeEle[1], "2305")
			}
		},
		func(response interface{}) {
			resp, ok := response.(market.SubscribeCandlestickResponse)
			if ok {
				if &resp != nil {
					if resp.Data != nil {
						applogger.Info("WebSocket returned data, count=%d", len(resp.Data))
						for _, t := range resp.Data {
							applogger.Info("Candlestick #%s data: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
								symbol, t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)

							tickerRetrive := &market.TickFloat{}
							err := client.Find(bson.M{"id": t.Id}).One(tickerRetrive)
							tickerWrite := t.TickToFloat()
							if err != nil {
								// if not exist, insert
								err = client.Insert(tickerWrite)
								if err != nil {
									applogger.Error("FlowWindowMarket: Failed to connection #%s db: %s", symbol, err.Error())
								} else {
									applogger.Info("FlowWindowMarket: Candlestick #%s data write to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										symbol, t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
								}
							} else {
								// if exist, update it for sync.
								// startTime should be equal to previousTick.Id
								selector := bson.M{"id": tickerWrite.Id}
								err := client.Update(selector, tickerWrite)
								if err != nil {
									applogger.Error("FlowWindowMarket: Failed to update #%s to db: %s", symbol, err.Error())
								} else {
									applogger.Info("FlowWindowMarket: Found Previous #%s Data, Update to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										symbol, t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
								}
							}
						}
					}
				}
			} else {
				applogger.Warn("Unknown response: %v", resp)
			}

		})

	wsClient.Connect(true)
	wsCandlestickClientMap[collectionName+"-flowWindowMarketInfo"] = wsClient
}

func makeTimeWindow(period periodUnit, startTime int64, toTime int64) ([][]int64, error) {
	var divisor int64
	var timeWindow [][]int64
	switch period {
	case Period_1min:
		divisor = 60
	case Period_5min:
		divisor = 300
	case Period_15min:
		divisor = 900
	case Period_30min:
		divisor = 1800
	case Period_60min:
		divisor = 3600
	case Period_4hour:
		divisor = 14400
	case Period_1day:
		divisor = 86400
	case Period_1week:
		divisor = 604800
	default:
		// month, year
		divisor = 0
		timeElement := []int64{startTime, toTime}
		timeWindow = append(timeWindow, timeElement)
		return timeWindow, nil
	}

	dataLength := (toTime - startTime) / divisor // Here the residual is less than divisor, such as 50s(60s), 50min(1h)...
	windowLength := dataLength / 300             // windowLength present how many slot of the period should be separated
	for i := 0; int64(i) < windowLength; i++ {
		start := startTime + int64(i)*divisor*300
		end := startTime + int64(i+1)*divisor*300
		timeElement := []int64{start, end}
		timeWindow = append(timeWindow, timeElement)
	}

	// add residual
	if (startTime + windowLength*divisor*300) < toTime {
		timeElement := []int64{startTime + windowLength*divisor*300, toTime}
		timeWindow = append(timeWindow, timeElement)
	}
	return timeWindow, nil
}

// -------------------------------------------------------------ORDER-------------------------------------------------------
func subOrderUpdateV2(symbol, accountID string) {
	// connect market db
	s, err := db.CreateRootDBSession()
	collectionName := fmt.Sprintf("HB-%s-%s", accountID, symbol)
	dbClient := s.DB("order").C(collectionName)
	mgoSessionMap[collectionName] = s
	if err != nil {
		applogger.Error("Failed to connection db: %s", err.Error())
		return
	}

	// seek accessKey with accountID
	accessKey, err := SeekAccountAccessKey(accountID)
	if err != nil {
		applogger.Error("subOrderUpdateV2: AccountMap could not found key matches the accountID %s", accountID)
		return
	}

	// Initialize a new instance
	wsClient := new(orderwebsocketclient.SubscribeOrderWebSocketV2Client).Init(
		config.GatewaySetting.GatewayHost,
		accessKey,
		config.HuoBiApiSetting.SecretKey,
		config.HuoBiApiSetting.ApiServerHost,
	)

	// Set the callback handlers
	wsClient.SetHandler(
		// Connected handler
		func(resp *auth.WebSocketV2AuthenticationResponse) {
			if resp.IsSuccess() {
				// Subscribe if authentication passed
				wsClient.Subscribe(symbol, "1149")
			} else {
				applogger.Error("Authentication error, code: %d, message:%s", resp.Code, resp.Message)
			}
		},
		// Response handler
		func(resp interface{}) {
			subResponse, ok := resp.(order.SubscribeOrderV2Response)
			if ok {
				if subResponse.Action == "sub" {
					if subResponse.IsSuccess() {
						applogger.Info("Subscription topic %s successfully", subResponse.Ch)
					} else {
						applogger.Error("Subscription topic %s error, code: %d, message: %s", subResponse.Ch, subResponse.Code, subResponse.Message)
					}
				} else if subResponse.Action == "push" {
					if subResponse.Data != nil {
						o := subResponse.Data
						oInDB := &order.OrderInfo{}
						err := dbClient.Find(bson.M{"orderid": o.OrderId}).One(oInDB)
						if err != nil {
							// not exist in db
							err = dbClient.Insert(o)
							if err != nil {
								applogger.Error("Failed to Insert data : %s", err.Error())
							} else {
								applogger.Info("Order created, event: %s, symbol: %s, type: %s, status: %s",
									o.EventType, o.Symbol, o.Type, o.OrderStatus)
							}
						} else {
							// update
							selector := bson.M{"orderid": o.OrderId}
							err := dbClient.Update(selector, o)
							if err != nil {
								applogger.Error("Failed to Update data : %s", err.Error())
							} else {
								applogger.Info("Order updated, event: %s, symbol: %s, type: %s, status: %s",
									o.EventType, o.Symbol, o.Type, o.OrderStatus)
							}
						}
					}
				}
			} else {
				applogger.Warn("Received unknown response: %v", resp)
			}
		})

	// Connect to the server and wait for the handler to handle the response
	wsClient.Connect(true)
	// HB-AccountID-Symbol
	wsOrderV2ClientMap[collectionName] = wsClient
}
