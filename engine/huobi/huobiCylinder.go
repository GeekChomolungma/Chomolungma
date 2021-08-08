package huobi

import (
	"strings"
	"time"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/db"
	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/marketwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/orderwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var mgoSessionMap map[string]*mgo.Session
var wsCandlestickClientMap map[string]*marketwebsocketclient.CandlestickWebSocketClient
var wsOrderV2ClientMap map[string]*orderwebsocketclient.SubscribeOrderWebSocketV2Client
var httpClientMap map[string]interface{}

type HuoBiCylinder struct {
}

// Ignite starts up to update data: market, account, order..
// Those data will be used internal Chmolungma
func (HBCylinder *HuoBiCylinder) Ignite() {
	mgoSessionMap = make(map[string]*mgo.Session)
	wsCandlestickClientMap = make(map[string]*marketwebsocketclient.CandlestickWebSocketClient)
	wsOrderV2ClientMap = make(map[string]*orderwebsocketclient.SubscribeOrderWebSocketV2Client)
	httpClientMap = make(map[string]interface{})

	// get symbols and write into DB
	querySymbolsAndWriteDisk()

	// subscribe the marketinfo
	// query sync time
	startTimeInt64, err := GetSyncStartTimestamp("HB-ethusdt-1min")
	if err != nil {
		applogger.Error("Ignite Huobi Server error: Can not connect mongodb for timestamp: %s", err.Error())
		return
	}

	endTime, err := GetTimestamp()
	if err != nil {
		applogger.Error("Ignite Huobi Server error: Can not get server timestamp: %s", err.Error())
		return
	}
	endTimeInt64 := int64(endTime + 60)
	flowWindowMarketInfo("ethusdt", Period_1min, startTimeInt64, endTimeInt64)
	subscribeMarketInfo("ethusdt", Period_1min)

	startTimeInt64, err = GetSyncStartTimestamp("HB-btcusdt-1min")
	if err != nil {
		applogger.Error("Ignite Huobi Server error: Can not connect mongodb for timestamp: %s", err.Error())
		return
	}
	flowWindowMarketInfo("btcusdt", Period_1min, startTimeInt64, endTimeInt64)
	subscribeMarketInfo("btcusdt", Period_1min)

	startTimeInt64, err = GetSyncStartTimestamp("HB-htusdt-1min")
	if err != nil {
		applogger.Error("Ignite Huobi Server error: Can not connect mongodb for timestamp: %s", err.Error())
		return
	}
	flowWindowMarketInfo("htusdt", Period_1min, startTimeInt64, endTimeInt64)
	subscribeMarketInfo("htusdt", Period_1min)

	for accountID, _ := range config.AccountMap {
		subOrderUpdateV2("btcusdt", accountID)
	}
}

func GetSyncStartTimestamp(collection string) (int64, error) {
	s, err := db.CreateMarketDBSession()
	if err != nil {
		return 0, err
	}
	startTimeInt64 := int64(1627805769)
	client := s.DB("marketinfo").C("HB-sync-timestamp")
	pst := &dtos.PreviousSyncTime{}
	err = client.Find(bson.M{"collectionname": collection}).One(pst)
	if err != nil {
		// not exist
	} else {
		// got it
		startTimeInt64 = pst.PreviousSyncTs
	}
	s.Close()
	return startTimeInt64, nil
}

// Flameout elegantly stop the Cylinder
func (HBCylinder *HuoBiCylinder) Flameout() {
	CandlestickClientFlameout() // market info
	OrderV2ClientFlameout()     // order info

	time.Sleep(time.Duration(2) * time.Second)
	for _, session := range mgoSessionMap {
		session.Close()
	}
}

func CandlestickClientFlameout() {
	// update previousTickMap into mongo
	s, err := db.CreateMarketDBSession()
	if err != nil {
		applogger.Error("CandlestickClientFlameout: HuoBi Flameout error Can not connect mongodb")
	}

	client := s.DB("marketinfo").C("HB-sync-timestamp")

	timeIteration := func(collection, timestamp interface{}) bool {
		prevSyncTime := &dtos.PreviousSyncTime{}
		err := client.Find(bson.M{"collectionname": collection}).One(prevSyncTime)
		prevSyncTime.CollectionName = collection.(string)
		prevSyncTime.PreviousSyncTs = timestamp.(int64)
		if err != nil {
			// not exist, insert flag
			err = client.Insert(prevSyncTime)
			if err != nil {
				applogger.Error("HuoBi Flameout, Insert %s sync time Error: %s", collection, err.Error())
			} else {
				applogger.Info("HuoBi Flameout, Inserted %s sync time: %d", collection, timestamp)
			}
		} else {
			//update
			selector := bson.M{"collectionname": collection}
			err := client.Update(selector, prevSyncTime)
			if err != nil {
				applogger.Error("HuoBi Flameout, Update %s sync time Error: %s", collection, err.Error())
			} else {
				applogger.Info("HuoBi Flameout, Updated %s sync time: %d", collection, timestamp)
			}
		}
		return true
	}
	PreviousSyncTimeMap.Range(timeIteration)
	s.Close()

	for collectionName, client := range wsCandlestickClientMap {
		sp := strings.Split(collectionName, "-")
		client.UnSubscribe(sp[1], sp[2], "2118")
		client.Close()
		applogger.Info("MarketInfo Client: %s closed", collectionName)
	}
}

func OrderV2ClientFlameout() {
	for collectionName, client := range wsOrderV2ClientMap {
		sp := strings.Split(collectionName, "-")
		client.UnSubscribe(sp[2], "1149")
		client.Close()
		applogger.Info("Order V2 Client: %s closed", collectionName)
	}
}
