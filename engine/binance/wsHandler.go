package binance

import (
	"context"
	"time"

	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	binance_connector "github.com/binance/binance-connector-go"
	"go.mongodb.org/mongo-driver/bson"
)

func SubscribeKlineStream(recordLabel, symbolName, intervalValue string) {
	metaCol := mongoInc.NewMetaCollection[*binance_connector.WsKlineEvent]("marketInfo", recordLabel, mongoInc.BinanKline)

	websocketStreamClient := binance_connector.NewWebsocketStreamClient(false)
	wsKlineHandler := func(event *binance_connector.WsKlineEvent) {
		eventStored := &binance_connector.WsKlineEvent{}
		lockMap[recordLabel].Lock()
		defer lockMap[recordLabel].Unlock()
		metaCol.Retrieve("kline.starttime", event.Kline.StartTime, eventStored)
		applogger.Debug("SubscribeKlineStream: Retrieve kline: %v", binance_connector.PrettyPrint(eventStored))
		if eventStored.Event != "kline" {
			// non exist, just insert it
			metaCol.Store("", event)
			applogger.Info("SubscribeKlineStream: insert the kline(%s) event: %v", recordLabel, binance_connector.PrettyPrint(event))
			return
		}

		if eventStored.Kline.IsFinal || eventStored.Kline.TradeNum > event.Kline.TradeNum {
			applogger.Warn("SubscribeKlineStream: kline(%s) starttime: %v has a later state(maybe finished synchron) than the incomming event, discard it.",
				recordLabel, eventStored.Kline.StartTime)
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

			stored := false
			for i := 0; i < 3; i++ {
				_, err := metaCol.Collection.UpdateOne(context.TODO(), filter, update)
				if err != nil {
					applogger.Error("SubscribeKlineStream: Update the kline(%s) startTime: %v, failed e: %v", recordLabel, event.Kline.StartTime, err)
					time.Sleep(time.Second)
					continue
				} else {
					applogger.Info("SubscribeKlineStream: Update the kline(%s) event: %v", recordLabel, binance_connector.PrettyPrint(event))
					stored = true
					break
				}
			}
			if !stored {
				applogger.Error("SubscribeKlineStream: UpdateOne the  kline(%s): %v, failed 3 times.", recordLabel, event.Kline.StartTime)
			}
		}
	}
	errHandler := func(err error) {
		applogger.Error("Symbol:%v, subscription error: %s", symbolName, err.Error())
		applogger.Info("Re Subscribe market info.")
		go SubscribeKlineStream(recordLabel, symbolName, intervalValue)
	}

	doneCh, _, err := websocketStreamClient.WsKlineServe(symbolName, intervalValue, wsKlineHandler, errHandler)
	if err != nil {
		applogger.Error("WsKlineServe error: %s", err.Error())
		return
	}
	<-doneCh
	applogger.Warn("Symbol: %v, WsKlineServe closed by doneCh", symbolName)
}
