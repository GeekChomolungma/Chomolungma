package main

import (
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/market"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	mongo, err := mgo.Dial("mongodb://market:admin123@localhost:27017")
	if err != nil {
		applogger.Error("Failed to connect to db: %s", err.Error())
		return
	}

	client := mongo.DB("marketinfo").C("HB-btcusdt-1min")
	case2(client)
}

func case1(client *mgo.Collection) {
	var t *market.Tick
	if t == nil {
		fmt.Println("t is nil", t)
	}

	tt := &market.Tick{}
	if tt == nil {
		fmt.Println("tt is nil", tt)
	}
	fmt.Println("tt is not nil", tt)

	tmp := &market.Tick{
		Id:    1626453720,
		Count: 1234,
	}

	// update the previous data
	selector := bson.M{"id": 1626453720}
	err := client.Update(selector, tmp)
	if err != nil {
		applogger.Error("Failed to update to db: %s", err.Error())
	}
}

func case2(client *mgo.Collection) {
	tick := &market.TickFloat{}
	iter := client.Find(nil).Sort("-id").Limit(10).Iter()
	for iter.Next(tick) {
		fmt.Println(tick.Id)
	}
}
