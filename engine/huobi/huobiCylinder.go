package huobi

import (
	"strings"
	"time"

	"github.com/GeekChomolungma/Chomolungma/db"
	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/clients/marketwebsocketclient"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	// query sync time
	startTimeInt64, err := GetSyncStartTimestamp("HB-btcusdt-1min")
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

	flowWindowMarketInfo("btcusdt", Period_1min, startTimeInt64, endTimeInt64)
	subscribeMarketInfo("btcusdt", Period_1min)
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
	CandlestickClientFlameout()

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
	for collection, timestamp := range PreviousSyncTimeMap {
		prevSyncTime := &dtos.PreviousSyncTime{}
		err := client.Find(bson.M{"collectionname": collection}).One(prevSyncTime)
		prevSyncTime.CollectionName = collection
		prevSyncTime.PreviousSyncTs = timestamp
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
	}
	s.Close()

	for collectionName, client := range wsCandlestickClientMap {
		sp := strings.Split(collectionName, "-")
		client.UnSubscribe(sp[0], string(sp[1]), "2118")
		client.Close()
		applogger.Info("Client: %s closed", collectionName)
	}
}
