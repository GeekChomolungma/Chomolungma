package huobi

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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

// subscribeMarketInfo records tickers, and the max tickers number is 300 once.
// So, startTime and toTime should be constrain for:
//                   (toTime - startTime) / period < 300
// if the inequality above not satisfy, a time window will be calculated, which
// splits the duration into multi 300-size slots.
func subscribeMarketInfo(label string) {
	collectionName := label
	var timeList []int64                         // TickMap auxiliary
	TickMap := make(map[int64]*market.TickFloat) // TickMap is a 10-size flow window
	rwMutex := new(sync.RWMutex)

	sp := strings.Split(label, "-") // label: HB-btcusdt-1min
	symbol := sp[1]
	period := periodUnit(sp[2])

	// connect market db
	s, err := db.CreateMarketDBSession()
	client := s.DB("marketinfo").C(collectionName)
	mgoSessionMap[collectionName] = s
	if err != nil {
		applogger.Error("subscribeMarketInfo: Failed to connection db, %s", err.Error())
		return
	}

	ntick := &market.TickFloat{}
	iter := client.Find(nil).Sort("-id").Limit(5).Iter()
	for iter.Next(ntick) {
		TickMap[ntick.Id] = ntick
	}

	// websocket
	wsClient := new(marketwebsocketclient.CandlestickWebSocketClient).Init(config.GatewaySetting.GatewayHost)
	wsClient.SetHandler(
		func() {
			applogger.Info("subscribeMarketInfo: HuoBi MarketInfo subscription #%s, K-period %s.", symbol, period)
			// caculate a sync time window
			timeWindow, err := makeTimeWindow(label, period)
			if err != nil {
				return
			}
			applogger.Info("subscribeMarketInfo: timeWindow length is %d, data is %d", len(timeWindow), timeWindow)
			wsClient.Subscribe(symbol, string(period), "1")

			// multi request for data returned under 300 once.
			// whatever the period is, request should make sure that
			// the res return data length less than 300.
			for _, timeEle := range timeWindow {
				wsClient.Request(symbol, string(period), timeEle[0], timeEle[1], "2")
			}
		},
		func(response interface{}) {
			resp, ok := response.(market.SubscribeCandlestickResponse)
			if ok {
				if &resp != nil {
					rwMutex.Lock()
					if resp.Tick != nil {
						t := resp.Tick
						if tick, exist := TickMap[t.Id]; !exist {
							// new time item OR too old item OR start/restart program
							// add current  tick into map
							// add previous tick into db
							tf := t.TickToFloat()
							TickMap[t.Id] = tf

							// sort the key, increasing
							for k := range TickMap {
								timeList = append(timeList, k)
							}
							sort.Slice(timeList, func(i, j int) bool {
								return timeList[i] < timeList[j]
							})

							if len(timeList) >= 2 {
								// after init db for a while
								// get previous tick
								previousTime := timeList[len(timeList)-2]
								previousTick := TickMap[previousTime]
								count, _ := client.Find(bson.M{"id": previousTick.Id}).Count()
								if count == 0 {
									// add previous tick into db
									previousTickWrite := previousTick
									err = client.Insert(previousTickWrite)
									if err != nil {
										applogger.Error("Failed to Insert #%s Tick: id: %d, count: %d, errmsg: %s",
											symbol, previousTick.Id, previousTick.Count, err.Error())
									}
									applogger.Info("New      #%s Tick Insert into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										symbol, previousTick.Id, previousTick.Count, previousTick.Vol,
										previousTick.Open, previousTick.Count, previousTick.Low, previousTick.High)
								} else {
									// Old time item received, which time is less than timeList bottom item's,
									// it suffered an long network delay(like 10*period), disregard it.
									// OR
									// Newest time Conflict with resp.Data received.
									// here may happen a competition with resp.Data, double inert!!!
									tickCmp := &market.TickFloat{}
									client.Find(bson.M{"id": previousTick.Id}).One(tickCmp) // must exist in db
									applogger.Error("Conflict #%s Tick(ts:%d, count:%d) has been inerted (ts:%d, count:%d) ,triggered by New Tick(ts:%d, count:%d).",
										symbol, previousTick.Id, previousTick.Count, tickCmp.Id, tickCmp.Count, tf.Id, tf.Count)
									if tickCmp.Count < previousTick.Count {
										selector := bson.M{"id": previousTick.Id}
										err := client.Update(selector, previousTick)
										if err != nil {
											applogger.Error("Failed to Update #%s Tick: id: %d, count: %d, errmsg: %s",
												symbol, previousTick.Id, previousTick.Count, err.Error())
										} else {
											applogger.Info("Previous #%s Tick Updated into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
												symbol, previousTick.Id, previousTick.Count, previousTick.Vol,
												previousTick.Open, previousTick.Count, previousTick.Low, previousTick.High)
										}
									}
								}
							}

							// remove the oldest item, which is 5th at the bottom
							if len(timeList) > 5 {
								// add PreviousSyncTime into map
								bottomTime := timeList[0]
								if SyncID, ok := PreviousSyncTimeMap.Load(collectionName); !ok {
									applogger.Error("Load PreviousSyncTimeMap failed: value for key: %s not exist.", collectionName)
									PreviousSyncTimeMap.Store(collectionName, TickMap[timeList[0]].Id)
								} else {
									SyncIDInt64 := SyncID.(int64)
									if bottomTime > SyncIDInt64 {
										PreviousSyncTimeMap.Store(collectionName, bottomTime)
									}
								}
								// only keep the 10 pass time items
								delete(TickMap, timeList[0])
								timeList = timeList[1:]
							}
						} else {
							// old tick received.
							// And this tick exists in map, update TickMap
							if t.Count <= tick.Count {
								// disregard the old tick in this time
								applogger.Error("Same time #%s Tick received (ts:%d, count:%d) , but Tick in Map is (ts:%d, count:%d), ignore it.",
									symbol, t.Id, t.Count, tick.Id, tick.Count)
							} else {
								// better tick, update TickMap
								tf := t.TickToFloat()
								TickMap[t.Id] = tf

								previousTime := timeList[len(timeList)-1]
								if tf.Id < previousTime {
									// if this tick not the top one, update into db
									// tf is in tick map, but not the newest tick
									tickCmp := &market.TickFloat{}
									client.Find(bson.M{"id": tf.Id}).One(tickCmp) // must exist in db
									if tickCmp.Count < tf.Count {
										selector := bson.M{"id": tf.Id}
										err := client.Update(selector, tf)
										if err != nil {
											applogger.Error("Failed to Update #%s Tick: id: %d, count: %d, errmsg: %s",
												symbol, tf.Id, tf.Count, err.Error())
										} else {
											applogger.Info("Previous #%s Tick Updated into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
												symbol,
												t.Id, t.Count, t.Vol,
												t.Open, t.Count, t.Low, t.High)
										}
									}
								}
							}
						}
					}

					// The history Tick data,
					// which are requested from startTime to toTime,
					// are included in resp.Data
					if resp.Data != nil {
						applogger.Info("Sync MarketInfo: WebSocket returned data, count=%d", len(resp.Data))
						for _, t := range resp.Data {
							// applogger.Info("Sync MarketInfo: Candlestick #%s data: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
							// 	symbol, t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
							tickerRetrive := &market.TickFloat{}
							err := client.Find(bson.M{"id": t.Id}).One(tickerRetrive)
							tickerWrite := t.TickToFloat()
							if err != nil {
								// if not exist, insert
								err = client.Insert(tickerWrite)
								if err != nil {
									applogger.Error("Sync MarketInfo: Failed to connection #%s db: %s", symbol, err.Error())
								} else {
									applogger.Info("Sync MarketInfo: Candlestick #%s data write to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										symbol, t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
								}
							} else {
								// if exist, update it for sync.
								// startTime should be equal to previousTick.Id
								if tickerRetrive.Count < tickerWrite.Count {
									selector := bson.M{"id": tickerWrite.Id}
									err := client.Update(selector, tickerWrite)
									if err != nil {
										applogger.Error("Sync MarketInfo: Failed to update #%s to db: %s", symbol, err.Error())
									} else {
										applogger.Info("Sync MarketInfo: Found Previous #%s Data, Update to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
											symbol, t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
									}
								}
							}
						}
					}
					rwMutex.Unlock()
				}
			} else {
				applogger.Warn("subscribeMarketInfo: Unknown response: %v", resp)
			}
		})

	wsClient.Connect(true)
	wsCandlestickClientMap[collectionName] = wsClient
}

func makeTimeWindow(label string, period periodUnit) ([][]int64, error) {
	startTime, err := GetSyncStartTimestamp(label)
	if err != nil {
		applogger.Error("subscribeMarketInfo: makeTimeWindow error, Can not connect mongodb for timestamp: %s", err.Error())
		return nil, errors.New("")
	}
	prevToTime, err := GetTimestamp()
	if err != nil {
		applogger.Error("subscribeMarketInfo: makeTimeWindow error, Huobi Server error: Can not get server timestamp: %s", err.Error())
		return nil, errors.New("")
	}
	toTime := int64(prevToTime + 60)

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
