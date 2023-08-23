package binance

import (
	"sync"

	binance_connector "github.com/binance/binance-connector-go"
)

var lockMap map[string]*sync.Mutex

type BinanCylinder struct {
}

// Ignite starts up to update data: market, account, order..
// Those data will be used inner Chmolungma
func (BACylinder *BinanCylinder) Ignite() {
	lockMap = make(map[string]*sync.Mutex)

	if _, ok := lockMap["ETCUSDT"]; !ok {
		lockMap["ETCUSDT"] = &sync.Mutex{}
	}

	binance_connector.WebsocketKeepalive = true
	go SubscribeKlineStream("ETCUSDT", "1m")

	currentTime := ServerTime()
	SyncHistoricalKline("ETCUSDT", "1m", 0, currentTime.ServerTime)
}

func (BACylinder *BinanCylinder) Flush() {

}

// Flameout elegantly stop the Cylinder
func (BACylinder *BinanCylinder) Flameout() {

}
