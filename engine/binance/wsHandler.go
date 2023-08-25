package binance

import (
	"context"

	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	binance_connector "github.com/binance/binance-connector-go"
	"gopkg.in/mgo.v2/bson"
)

func SubscribeKlineStream(symbolName, intervalValue string) {
	metaCol := mongoInc.NewMetaCollection[*binance_connector.WsKlineEvent]("marketInfo", symbolName, mongoInc.BinanKline)

	websocketStreamClient := binance_connector.NewWebsocketStreamClient(false)
	wsKlineHandler := func(event *binance_connector.WsKlineEvent) {
		eventStored := &binance_connector.WsKlineEvent{}
		lockMap[symbolName].Lock()
		defer lockMap[symbolName].Unlock()
		metaCol.Retrieve("kline.starttime", event.Kline.StartTime, eventStored)
		applogger.Debug("SubscribeKlineStream: Retrieve kline: %v", binance_connector.PrettyPrint(eventStored))
		if eventStored.Event != "kline" {
			// non exist, just insert it
			metaCol.Store("", event)
			applogger.Info("SubscribeKlineStream: insert the kline event: %v", binance_connector.PrettyPrint(event))
			return
		}

		if eventStored.Kline.IsFinal || eventStored.Kline.TradeNum > event.Kline.TradeNum {
			applogger.Warn("SubscribeKlineStream: kline.starttime: %v has a later state(maybe finished synchron) than the incomming event, discard it.",
				eventStored.Kline.StartTime)
			return
		} else {
			// update:
			// LastTradeID          int64  `json:"L"`
			// Open                 string `json:"o"`
			// Close                string `json:"c"`
			// High                 string `json:"h"`
			// Low                  string `json:"l"`
			// Volume               string `json:"v"`
			// TradeNum             int64  `json:"n"`
			// IsFinal              bool   `json:"x"`
			// QuoteVolume          string `json:"q"`
			// ActiveBuyVolume      string `json:"V"`
			// ActiveBuyQuoteVolume string `json:"Q"`

			filter := bson.D{{"kline.starttime", event.Kline.StartTime}}
			update := bson.D{{"$set",
				bson.D{
					{"kline.lasttradeid", event.Kline.LastTradeID},
					{"kline.open", event.Kline.Open},
					{"kline.close", event.Kline.Close},
					{"kline.high", event.Kline.High},
					{"kline.low", event.Kline.Low},
					{"kline.volume", event.Kline.Volume},
					{"kline.tradenum", event.Kline.TradeNum},
					{"kline.isfinal", event.Kline.IsFinal},
					{"kline.quotevolume", event.Kline.QuoteVolume},
					{"kline.activebuyvolume", event.Kline.ActiveBuyVolume},
					{"kline.activebuyquotevolume", event.Kline.ActiveBuyQuoteVolume},
				}}}
			metaCol.Collection.UpdateOne(context.TODO(), filter, update)
			applogger.Info("SubscribeKlineStream: update the kline event: %v", binance_connector.PrettyPrint(event))
		}
	}
	errHandler := func(err error) {
		applogger.Error("Symbol:%v, subscription error: %s", symbolName, err.Error())
		applogger.Info("Re Subscribe market info.")
		go SubscribeKlineStream(symbolName, intervalValue)
	}

	doneCh, _, err := websocketStreamClient.WsKlineServe(symbolName, intervalValue, wsKlineHandler, errHandler)
	if err != nil {
		applogger.Error("WsKlineServe error: %s", err.Error())
		return
	}
	<-doneCh
	applogger.Warn("Symbol: %v, WsKlineServe closed by doneCh", symbolName)
}
