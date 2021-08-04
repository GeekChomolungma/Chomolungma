package huobi

import (
	"strings"
	"time"

	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/marketwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2"
)

var mgoSessionMap map[string]*mgo.Session
var wsCandlestickClientMap map[string]*marketwebsocketclient.CandlestickWebSocketClient
var httpClientMap map[string]interface{}

type HuoBiCylinder struct {
}

// Ignite starts up to update data: market, account, order..
// Those data will be used internal Chmolungma
func (HBCylinder *HuoBiCylinder) Ignite() {
	mgoSessionMap = make(map[string]*mgo.Session)
	wsCandlestickClientMap = make(map[string]*marketwebsocketclient.CandlestickWebSocketClient)
	httpClientMap = make(map[string]interface{})

	// get symbols and write into DB
	querySymbolsAndWriteDisk()

	// subscribe the marketinfo
	flowWindowMarketInfo("btcusdt", Period_1min, 1627896420, 1628059200)
	subscribeMarketInfo("btcusdt", Period_1min)
}

// Flameout elegantly stop the Cylinder
func (HBCylinder *HuoBiCylinder) Flameout() {
	CandlestickClientFlameout()

	time.Sleep(time.Duration(2) * time.Second)
	for _, session := range mgoSessionMap {
		session.Close()
	}
}

func CandlestickClientFlameout() {
	for collectionName, client := range wsCandlestickClientMap {
		sp := strings.Split(collectionName, "-")
		client.UnSubscribe(sp[0], string(sp[1]), "2118")
		client.Close()
		applogger.Info("Client: %s closed", collectionName)
	}
}
