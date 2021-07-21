package huobi

import (
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/db"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/marketwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/market"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2/bson"
)

var tmp *market.Tick

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

func subscribeMarketInfo(symbol string, period periodUnit) {
	// connect market db
	s, err := db.CreateMarketDBSession()
	collectionName := fmt.Sprintf("%s-%s", symbol, string(period))
	client := s.DB("marketinfo").C(collectionName)
	mgoSessionMap[collectionName+"-subscribeMarketInfo"] = s
	if err != nil {
		applogger.Error("Failed to connection db: %s", err.Error())
		return
	}

	// websocket
	wsClient := new(marketwebsocketclient.CandlestickWebSocketClient).Init(config.GatewaySetting.GatewayHost)

	wsClient.SetHandler(
		func() {
			//wsClient.Request(symbol, string(period), 1569361140, 1569366420, "2305")
			wsClient.Subscribe(symbol, string(period), "2118")
		},
		func(response interface{}) {
			resp, ok := response.(market.SubscribeCandlestickResponse)
			if ok {
				if &resp != nil {
					if resp.Tick != nil {
						t := resp.Tick
						applogger.Info("Candlestick update, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
							t.Id, t.Count, t.Vol, t.Open, t.Close, t.Low, t.High)

						ticker := &market.Tick{}
						err := client.Find(bson.M{"id": t.Id}).One(ticker)
						if err != nil {
							// if not exist, insert
							// insert new data
							applogger.Info("Failed to find ID in db, insert the new data.")
							err = client.Insert(t)
							if err != nil {
								applogger.Error("Failed to Insert data : %s", err.Error())
							} else {
								applogger.Info("Write to db success, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
									t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
							}

							if tmp != nil {
								applogger.Info("Previous Data is: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
									tmp.Id, tmp.Count, tmp.Vol, tmp.Open, tmp.Count, tmp.Low, tmp.High)

								// update the previous data
								selector := bson.M{"id": tmp.Id}
								err := client.Update(selector, tmp)
								if err != nil {
									applogger.Error("Failed to update to db: %s", err.Error())
								} else {
									applogger.Info("Candlestick update to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										tmp.Id, tmp.Count, tmp.Vol, tmp.Open, tmp.Close, tmp.Low, tmp.High)
								}
							}
						}
						tmp = t
					}

					if resp.Data != nil {
						applogger.Info("WebSocket returned data, count=%d", len(resp.Data))
						for _, t := range resp.Data {
							applogger.Info("Candlestick data, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
								t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)

							ticker := &market.Tick{}
							err := client.Find(bson.M{"id": t.Id}).One(ticker)
							if err != nil {
								// if not exist, insert
								applogger.Error("not exist, insert")
								err = client.Insert(&t)
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
	collectionName := fmt.Sprintf("%s-%s", symbol, string(period))
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
							applogger.Info("Candlestick data, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
								t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)

							ticker := &market.Tick{}
							err := client.Find(bson.M{"id": t.Id}).One(ticker)
							if err != nil {
								// if not exist, insert
								applogger.Error("not exist, insert")
								err = client.Insert(&t)
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
