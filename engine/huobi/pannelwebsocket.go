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

func subscribeMarketInfo() {
	// connect market db
	s, err := db.CreateMarketDBSession()
	client := s.DB("marketinfo").C("btcusdt")
	defer s.Close()
	if err != nil {
		applogger.Error("Failed to connection db: %s", err.Error())
		return
	}

	// websocket
	wsClient := new(marketwebsocketclient.CandlestickWebSocketClient).Init(config.GatewaySetting.GatewayHost)

	wsClient.SetHandler(
		func() {
			wsClient.Request("btcusdt", "1min", 1569361140, 1569366420, "2305")
			wsClient.Subscribe("btcusdt", "1min", "2118")
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
								applogger.Error("Failed to find ID in db: %s", err.Error())
								continue
							}
							if ticker.Id == t.Id {
								continue
							}
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
			} else {
				applogger.Warn("Unknown response: %v", resp)
			}

		})

	wsClient.Connect(true)

	fmt.Println("Press ENTER to unsubscribe and stop...")
	fmt.Scanln()

	wsClient.UnSubscribe("btcusdt", "1min", "2118")

	wsClient.Close()
	applogger.Info("Client closed")
}
