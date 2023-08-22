package binance

import (
	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	binance_connector "github.com/binance/binance-connector-go"
)

func SubscribeKlineStream(symbolName, intervalValue string) {
	metaCol := mongoInc.NewMetaCollection[*binance_connector.WsKlineEvent]("marketInfo", symbolName, mongoInc.BinanKline)

	websocketStreamClient := binance_connector.NewWebsocketStreamClient(false)
	wsKlineHandler := func(event *binance_connector.WsKlineEvent) {
		metaCol.Store("", event)
		applogger.Info(binance_connector.PrettyPrint(event))
	}
	errHandler := func(err error) {
		applogger.Error("%v subscription error: %s", symbolName, err.Error())
	}

	doneCh, _, err := websocketStreamClient.WsKlineServe(symbolName, intervalValue, wsKlineHandler, errHandler)
	if err != nil {
		applogger.Error("WsKlineServe error: %s", err.Error())
		return
	}
	<-doneCh
	applogger.Warn("WsKlineServe closed by doneCh")
}
