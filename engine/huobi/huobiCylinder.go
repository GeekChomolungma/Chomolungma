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
var FlushDuration int

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
	for _, label := range config.HBMarketSubList {
		subscribeMarketInfo(label)
	}

	for accountID := range config.AccountMap {
		subOrderUpdateV2("btcusdt", accountID)
	}
}

func GetSyncStartTimestamp(collection string) int64 {
	// in PreviousSyncTimeMap
	if startTime, ok := PreviousSyncTimeMap.Load(collection); ok {
		return startTime.(int64)
	}

	// 2021-08-01 00:00:00, day, week, month tick is consistent here.
	startTimeInt64 := int64(1627747200)

	// in db
	s, err := db.CreateMarketDBSession()
	if err != nil {
		return startTimeInt64
	}
	client := s.DB("marketinfo").C("HB-sync-timestamp")
	pst := &dtos.PreviousSyncTime{}
	err = client.Find(bson.M{"collectionname": collection}).One(pst)
	if err == nil {
		// got it
		startTimeInt64 = pst.PreviousSyncTs
	}
	s.Close()
	PreviousSyncTimeMap.Store(collection, startTimeInt64)
	return startTimeInt64
}

func (HBCylinder *HuoBiCylinder) Flush() {
	FlushDuration = 3600
	go flushPerSecond(FlushDuration)
}

func flushPerSecond(sec int) {
	// every sec flush sync timestamp
	ticker := time.NewTicker(time.Duration(sec) * time.Second)
	for {
		select {
		case <-ticker.C:
			applogger.Debug("Flush: HuoBi MarketInfo flush ticker time up, call flushSyncTime.")
			flushSyncTime()
		}
	}
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
	flushSyncTime()

	for collectionName, client := range wsCandlestickClientMap {
		sp := strings.Split(collectionName, "-")
		client.UnSubscribe(sp[1], sp[2], "2118")
		client.Close()
		applogger.Info("MarketInfo Client: %s closed", collectionName)
	}
}

func flushSyncTime() {
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
				applogger.Error("HuoBi flushSyncTime, Insert %s sync time Error: %s", collection, err.Error())
			} else {
				applogger.Info("HuoBi flushSyncTime, Inserted %s sync time: %d", collection, timestamp)
			}
		} else {
			//update
			selector := bson.M{"collectionname": collection}
			err := client.Update(selector, prevSyncTime)
			if err != nil {
				applogger.Error("HuoBi flushSyncTime, Update %s sync time Error: %s", collection, err.Error())
			} else {
				applogger.Info("HuoBi flushSyncTime, Updated %s sync time: %d", collection, timestamp)
			}
		}
		return true
	}
	PreviousSyncTimeMap.Range(timeIteration)
	s.Close()
}

func OrderV2ClientFlameout() {
	for collectionName, client := range wsOrderV2ClientMap {
		sp := strings.Split(collectionName, "-")
		client.UnSubscribe(sp[2], "1149")
		client.Close()
		applogger.Info("Order V2 Client: %s closed", collectionName)
	}
}
