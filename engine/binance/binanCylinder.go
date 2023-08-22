package binance

import (
	binance_connector "github.com/binance/binance-connector-go"
)

type BinanCylinder struct {
}

// Ignite starts up to update data: market, account, order..
// Those data will be used inner Chmolungma
func (BACylinder *BinanCylinder) Ignite() {
	binance_connector.WebsocketKeepalive = true
	go SubscribeKlineStream("ETCUSDT", "1m")

	currentTime := ServerTime()
	SyncHistoricalKline("ETCUSDT", "1m", 1692670260000, currentTime.ServerTime)
}

func (BACylinder *BinanCylinder) Flush() {

}

// Flameout elegantly stop the Cylinder
func (BACylinder *BinanCylinder) Flameout() {

}
