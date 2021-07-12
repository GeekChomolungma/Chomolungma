package huobi

import (
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/marketwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/market"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
)

func subscribeMarketInfo() {
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
					}

					if resp.Data != nil {
						applogger.Info("WebSocket returned data, count=%d", len(resp.Data))
						for _, t := range resp.Data {
							applogger.Info("Candlestick data, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
								t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
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
