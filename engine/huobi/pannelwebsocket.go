package huobi

import (
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
	var endTs int64
	DataAlreadySync := false
	subStart := false
	var timeList []int64 // TickMap auxiliary

	// TickMap is a const size(like 5) flow window, to cache new tick received from remote,
	// it can reduce pressure of DB read and write.
	TickMap := make(map[int64]market.TickFloat)

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

	rwMutex := new(sync.RWMutex)
	// websocket
	wsClient := new(marketwebsocketclient.CandlestickWebSocketClient).Init(config.GatewaySetting.GatewayHost)
	wsClient.SetHandler(
		func() {
			subStart = false

			// init TickMap
			ntick := market.TickFloat{}
			iter := client.Find(nil).Sort("-id").Limit(5).Iter()
			for iter.Next(&ntick) {
				TickMap[ntick.Id] = ntick
			}

			applogger.Info("subscribeMarketInfo: HuoBi MarketInfo subscription #%s, K-period %s.", symbol, period)
			wsClient.Subscribe(symbol, string(period), "1")
		},

		func(response interface{}) {
			resp, ok := response.(market.SubscribeCandlestickResponse)
			if ok {
				if &resp != nil {
					if resp.Tick != nil {
						t := resp.Tick
						// start sync once
						if !subStart {
							subStart = true
							endTs = t.Id
							applogger.Info("Sync MarketInfo: #%s-%s sync start, current timestamp is %d", symbol, period, endTs)

							// caculate a sync time window
							startTime := GetSyncStartTimestamp(label)
							DataAlreadySync = syncHistory(label, startTime, endTs, wsClient)
						}

						if tick, exist := TickMap[t.Id]; !exist {
							// new time item OR too old item OR start/restart program
							// add current  tick into map
							// add previous tick into db
							tf := t.TickToFloat()
							TickMap[t.Id] = *tf

							// sort the key, increasing
							timeList = nil
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
								tickCmp := &market.TickFloat{}
								err := client.Find(bson.M{"id": previousTick.Id}).One(tickCmp)
								if err != nil {
									// add previous tick into db
									previousTickWrite := previousTick
									err = client.Insert(previousTickWrite)
									if err != nil {
										applogger.Error("Failed to Insert #%s-%s Tick: id: %d, count: %d, errmsg: %s",
											symbol, period, previousTick.Id, previousTick.Count, err.Error())
									}
									applogger.Info("New      #%s-%s Tick Insert into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
										symbol, period, previousTick.Id, previousTick.Count, previousTick.Vol,
										previousTick.Open, previousTick.High, previousTick.Low, previousTick.Close)
								} else {
									// LESS CHANCE HAPPEN: fixed the history write disk logic, no dirty from history action, so this will not happen.
									// Old time item received, which time is less than timeList bottom item's,
									// it suffered an long network delay(like 10*period), disregard it.
									// OR
									// Newest time Conflict with resp.Data received.
									// here may happen a competition with resp.Data, double inert!!!
									// OR
									// restart and reload TickMap
									applogger.Info("Conflict #%s-%s Tick(ts:%d, count:%d) in map (timelist: %v) has been inserted in DB(ts:%d, count:%d),triggered by New Tick(ts:%d, count:%d).",
										symbol, period, previousTick.Id, previousTick.Count, timeList, tickCmp.Id, tickCmp.Count, tf.Id, tf.Count)
									if tickCmp.Count < previousTick.Count {
										selector := bson.M{"id": previousTick.Id}
										err := client.Update(selector, previousTick)
										if err != nil {
											applogger.Error("Failed to Update #%s-%s Tick: id: %d, count: %d, errmsg: %s",
												symbol, period, previousTick.Id, previousTick.Count, err.Error())
										} else {
											applogger.Info("Previous #%s-%s Tick Updated into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
												symbol, period, previousTick.Id, previousTick.Count, previousTick.Vol,
												previousTick.Open, previousTick.High, previousTick.Low, previousTick.Close)
										}
									}
								}
							}

							// remove the oldest item, which is 6th at the bottom
							if len(timeList) > 5 {
								// add PreviousSyncTime into map
								// TODO: should confirm all resp.Data are synced, else the unsynced part will lose.
								if DataAlreadySync {
									bottomTime := timeList[0]
									if SyncID, ok := PreviousSyncTimeMap.Load(collectionName); !ok {
										PreviousSyncTimeMap.Store(collectionName, bottomTime)
										applogger.Error("Load PreviousSyncTimeMap failed: value for key: %s not exist. store bottom %d in", collectionName, bottomTime)
									} else {
										SyncIDInt64 := SyncID.(int64)
										if bottomTime > SyncIDInt64 {
											PreviousSyncTimeMap.Store(collectionName, bottomTime)
											applogger.Debug("Store PreviousSyncTimeMap: key-%s, value-%d.", collectionName, bottomTime)
										}
									}
								}

								// only keep the 5 pass time items
								delete(TickMap, timeList[0])
							}
						} else {
							// old tick received.
							// And this tick exists in map, update TickMap
							if t.Count <= tick.Count {
								// disregard the old tick in this time
								applogger.Info("Same time #%s-%s Tick received (ts:%d, count:%d) , but Tick in Map is (ts:%d, count:%d), ignore it.",
									symbol, period, t.Id, t.Count, tick.Id, tick.Count)
							} else {
								// better tick, update TickMap
								// mostly, tick is the newest in tickmap
								tf := t.TickToFloat()
								TickMap[t.Id] = *tf
								applogger.Info("Same time #%s-%s Tick received (ts:%d, count:%d) , is better than Tick in Map is (ts:%d, count:%d), update tick map.",
									symbol, period, t.Id, t.Count, tick.Id, tick.Count)

								currentTime := timeList[len(timeList)-1]
								if tf.Id < currentTime {
									// LESS CHANCE HAPPEN: Old tick received because of impossible network time delay,
									//                     or when the newest tick created but the last tick suffer a small delay in net.
									// if this tick not the top one, update into db
									// tf is in tick map, but not the newest tick
									applogger.Error("Old time #%s-%s Tick received (ts:%d, count:%d), but Tick in Map is (ts:%d, count:%d), and currentTime in TickMap is (ts:%d).",
										symbol, period, tf.Id, tf.Count, tick.Id, tick.Count, currentTime)
									tickCmp := &market.TickFloat{}
									client.Find(bson.M{"id": tf.Id}).One(tickCmp) // must exist in db
									if tickCmp.Count < tf.Count {
										selector := bson.M{"id": tf.Id}
										err := client.Update(selector, tf)
										if err != nil {
											applogger.Error("Failed to Update #%s-%s Tick: id: %d, count: %d, errmsg: %s",
												symbol, period, tf.Id, tf.Count, err.Error())
										} else {
											applogger.Info("Previous #%s-%s Tick Updated into DB: id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
												symbol, period,
												t.Id, t.Count, t.Vol,
												t.Open, t.High, t.Low, t.Close)
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
						go func() {
							rwMutex.Lock()
							applogger.Info("Sync MarketInfo: WebSocket returned #%s-%s data, count: %d", symbol, period, len(resp.Data))
							for _, t := range resp.Data {
								tf := t.TickToFloat()
								if tf.Id >= endTs {
									// over endTs means the data is endTs
									// it's not history but in subscribing now.
									applogger.Info("Sync MarketInfo: #%s-%s history tick received(ts: %d) is later than or equal to the first sub tick(ts: %d), skip it.",
										symbol, period, tf.Id, endTs)
									continue
								}
								applogger.Info("Sync MarketInfo: #%s-%s history tick received(ts: %d) is earlier than the first sub tick(ts: %d), try to insert it.",
									symbol, period, tf.Id, endTs)

								tickCmp := &market.TickFloat{}
								err := client.Find(bson.M{"id": tf.Id}).One(tickCmp)
								if err != nil {
									// if not exist, insert
									err = client.Insert(tf)
									if err != nil {
										applogger.Error("Sync MarketInfo: Failed to connection #%s-%s db: %s", symbol, period, err.Error())
									} else {
										applogger.Info("Sync MarketInfo: Candlestick #%s-%s data write to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
											symbol, period, tf.Id, tf.Count, tf.Vol, tf.Open, tf.High, tf.Low, tf.Close)
									}
								} else {
									// LESS CHANCE HAPPEN: when restart, reconnect and query history data with sync time flag.
									// if exist, update it for sync.
									// startTime should be equal to previousTick.Id
									if tickCmp.Count < tf.Count {
										selector := bson.M{"id": tf.Id}
										err := client.Update(selector, tf)
										if err != nil {
											applogger.Error("Sync MarketInfo: Failed to update #%s-%s to db: %s", symbol, period, err.Error())
										} else {
											applogger.Info("Sync MarketInfo: Found Previous #%s-%s Data, Update to db, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
												symbol, period, tf.Id, tf.Count, tf.Vol, tf.Open, tf.High, tf.Low, tf.Close)
										}
									}
								}
							}

							sort.Slice(resp.Data, func(i, j int) bool {
								return resp.Data[i].Id < resp.Data[j].Id
							})

							DataAlreadySync = syncHistory(label, resp.Data[len(resp.Data)-1].Id, endTs, wsClient)
							if DataAlreadySync {
								applogger.Info("Sync MarketInfo: #%s-%s all history data synced finished, update ts: %d into PreviousSyncTimeMap.", symbol, period, resp.Data[len(resp.Data)-1].Id)
							}
							rwMutex.Unlock()
						}()
					}
				}
			} else {
				applogger.Warn("subscribeMarketInfo: Unknown response: %v", resp)
			}
		})

	wsClient.Connect(true)
	wsCandlestickClientMap[collectionName] = wsClient
}

func syncHistory(label string, startTime, endTs int64, wsClient *marketwebsocketclient.CandlestickWebSocketClient) bool {
	sp := strings.Split(label, "-") // label: HB-btcusdt-1min
	symbol := sp[1]
	period := periodUnit(sp[2])
	timeWindow, _ := calcuFirstTimeWindow(label, startTime, endTs)
	alreadySync := false
	if timeWindow == nil {
		alreadySync = true
	} else {
		// multi request for data returned under 300 once.
		// whatever the period is, request should make sure that
		// the res return data length less than 300.
		wsClient.Request(symbol, string(period), timeWindow[0], timeWindow[1], "2")
		applogger.Info("Sync MarketInfo: #%s-%s sync req sent, start: %d, end: %d, but from:%d, to:%d", symbol, period, startTime, endTs, timeWindow[0], timeWindow[1])
	}
	return alreadySync
}

func calcuFirstTimeWindow(label string, startTs, endTs int64) ([]int64, int) {
	sp := strings.Split(label, "-") // label: HB-btcusdt-1min
	symbol := sp[1]
	period := periodUnit(sp[2])
	var divisor int64
	var timeWindow []int64
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
	}

	if startTs != int64(1627747200) {
		// db has inited
		// if startTs is 1627747200, it means db does not have that init data.
		startTs = startTs + divisor
	}

	if (startTs + divisor) >= endTs {
		applogger.Info("subscribeMarketInfo: #%s-%s sync finish data length is %d, start:%d, to:%d.",
			symbol, period, len(timeWindow), startTs, endTs)
		return nil, 0
	}

	tsResidual := endTs % divisor
	if tsResidual != 0 {
		applogger.Error("subscribeMarketInfo: #%s-%s makeTimeWindowWithEndTs error, received unmatchable end timestamp for sync: %d", symbol, period, endTs)
	}

	dataLength := (endTs-startTs)/divisor + 1
	if dataLength > 300 {
		dataLength = 300
		timeWindow = []int64{startTs, startTs + 299*divisor}
	} else {
		timeWindow = []int64{startTs, endTs}
	}

	applogger.Info("subscribeMarketInfo: #%s-%s timeWindow length is %d, start:%d, to:%d, datalength is %d. time window is %v",
		symbol, period, len(timeWindow), timeWindow[0], timeWindow[1], dataLength, timeWindow)
	return timeWindow, int(dataLength)
}

func timeWindowAtEndTime(label string, period periodUnit, startTime, endTime int64) ([][]int64, int, error) {
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
		timeElement := []int64{startTime, endTime}
		timeWindow = append(timeWindow, timeElement)
		return timeWindow, 0, nil
	}

	dataLength := ((endTime - startTime) / divisor) + 1
	windowLength := dataLength / 300
	for i := 0; int64(i) < windowLength; i++ {
		start := startTime + int64(i)*300*divisor           // 0                  60*300
		end := startTime + int64(i+1)*300*divisor - divisor // 60*300             60*300 + 60*300
		timeElement := []int64{start, end}                  // [0:60:60*300)      [60*300:60:60*300*2)
		timeWindow = append(timeWindow, timeElement)        // [) [) [) [) [)
	}

	residual := dataLength % 300
	if residual > 0 {
		startResidual := startTime + windowLength*300*divisor
		timeElementResidual := []int64{startResidual, endTime}
		timeWindow = append(timeWindow, timeElementResidual)
	}
	return timeWindow, int(dataLength), nil
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

	// seek secretKey with accountID
	secretKey, err := SeekAccountSecretKey(accountID)
	if err != nil {
		applogger.Error("subOrderUpdateV2: SecretMap could not found key matches the accountID %s", accountID)
		return
	}

	// Initialize a new instance
	wsClient := new(orderwebsocketclient.SubscribeOrderWebSocketV2Client).Init(
		config.GatewaySetting.GatewayHost,
		accessKey,
		secretKey,
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
