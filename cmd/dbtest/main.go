package main

import (
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/market"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {

	tmp := &market.Tick{
		Id:    1626430680,
		Count: 9999,
	}

	mongo, err := mgo.Dial("mongodb://market:admin123@localhost:27017")
	if err != nil {
		applogger.Error("Failed to connect to db: %s", err.Error())
		return
	}

	client := mongo.DB("marketinfo").C("btcusdt")
	//ticker := &market.Tick{}
	//err = client.Find(bson.M{"id": 1626430680}).One(ticker)
	// if not exist, insert
	//applogger.Error("Failed to find ID in db: %s", err.Error())

	// update the previous data
	selector := bson.M{"id": 1626430680}
	err = client.Update(selector, tmp)
	if err != nil {
		applogger.Error("Failed to update to db: %s", err.Error())
	}

}
