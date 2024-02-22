package binance

import (
	"sync"

	"github.com/GeekChomolungma/Chomolungma/config"
	binance_connector "github.com/binance/binance-connector-go"
)

var lockMap map[string]*sync.Mutex

type BinanCylinder struct {
}

// Ignite starts up to update data: market, account, order..
// Those data will be used inner Chmolungma
func (BACylinder *BinanCylinder) Ignite() {
	lockMap = make(map[string]*sync.Mutex)
	binance_connector.WebsocketKeepalive = true
	for _, marketUnit := range config.BinanceMarketSubList {
		lockMap[marketUnit.RecordLabel] = &sync.Mutex{}
		go SubscribeKlineStream(marketUnit.RecordLabel, marketUnit.Symbol, marketUnit.Interval) //"Binance-ETCUSDT-1m", "ETCUSDT", "1m"

		currentTime := ServerTime()
		SyncHistoricalKline(marketUnit.RecordLabel, marketUnit.Symbol, marketUnit.Interval, 0, currentTime.ServerTime) // "ETCUSDT", "1m"
	}
}

func (BACylinder *BinanCylinder) Flush() {

}

// Flameout elegantly stop the Cylinder
func (BACylinder *BinanCylinder) Flameout() {

}
