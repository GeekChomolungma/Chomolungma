package binance

import (
	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	binance_connector "github.com/binance/binance-connector-go"
)

type BinanCylinder struct {
}

// Ignite starts up to update data: market, account, order..
// Those data will be used inner Chmolungma
func (BACylinder *BinanCylinder) Ignite() {
	go func() {
		metaCol := mongoInc.NewMetaCollection[*binance_connector.WsKlineEvent]("marketInfo", "BTCUSDT", mongoInc.BinanKline)

		websocketStreamClient := binance_connector.NewWebsocketStreamClient(false)
		wsKlineHandler := func(event *binance_connector.WsKlineEvent) {
			metaCol.Store("", event)
			applogger.Info(binance_connector.PrettyPrint(event))
		}
		errHandler := func(err error) {
			applogger.Error("BTCUSDT subscription error: %s", err.Error())
		}
		doneCh, _, err := websocketStreamClient.WsKlineServe("BTCUSDT", "1m", wsKlineHandler, errHandler)
		if err != nil {
			applogger.Error("WsKlineServe error: %s", err.Error())
			return
		}
		<-doneCh
	}()
}

func (BACylinder *BinanCylinder) Flush() {

}

// Flameout elegantly stop the Cylinder
func (BACylinder *BinanCylinder) Flameout() {

}
