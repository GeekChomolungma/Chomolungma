package binance

import (
	"context"

	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	binance_connector "github.com/binance/binance-connector-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	baseURL         = "https://api.binance.com"
	TEST_START_TIME = 1640966400000
)

// ServerTime gets binance time
// use it as the end time for kline/ticker sync
func ServerTime() *binance_connector.ServerTimeResponse {

	client := binance_connector.NewClient("", "") // use the default URL

	// set to debug mode
	client.Debug = true

	// NewServerTimeService
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		applogger.Error("Sync ServerTime failed, e: %v", err)
		return nil
	}

	return serverTime
}

type SyncFlag struct {
	Symbol    string
	StartTime uint64
}

func SyncHistoricalKline(recordLabel, symbolName, intervalValue string, startTime, endTime uint64, hsFlag *historySynced) {
	syncFlagCol := mongoInc.NewMetaCollection[*SyncFlag]("marketSyncFlag", recordLabel, mongoInc.BinanSyncFlag)
	syncFlag := &SyncFlag{}
	if startTime < TEST_START_TIME {
		syncFlagCol.Retrieve("symbol", symbolName, syncFlag)
		if syncFlag.StartTime < TEST_START_TIME {
			syncFlag.StartTime = TEST_START_TIME
		}
		startTime = syncFlag.StartTime
	}

	// repeated time between (n-1)th window's end and nth window's begin
	// to fix
	timeWindow := calcTimeWindow(periodUnit(intervalValue), startTime, endTime)

	client := binance_connector.NewClient("", "", baseURL)

	for k, timePair := range timeWindow {
		// Klines
		klines, err := client.NewKlinesService().
			Symbol(symbolName).Interval(intervalValue).StartTime(timePair[0]).EndTime(timePair[1]).Do(context.Background())
		if err != nil {
			applogger.Warn("SyncHistoricalKline: GetKline(%s) failed once: timeWindow(500) index is %d, start:%d, to:%d", recordLabel, k, timePair[0], timePair[1])
			continue
		} else {
			applogger.Info("SyncHistoricalKline: GetKline(%s) succeeded once: timeWindow(500) index is %d, start:%d, to:%d, kline count: %v", recordLabel, k, timePair[0], timePair[1], len(klines))
		}

		metaCol := mongoInc.NewMetaCollection[*binance_connector.WsKlineEvent]("marketInfo", recordLabel, mongoInc.BinanKline)
		for _, kline := range klines {
			eventStored := &binance_connector.WsKlineEvent{}
			metaCol.Retrieve("kline.starttime", kline.OpenTime, eventStored)
			if eventStored.Event != "kline" {
				// non exist, just insert it
				klineStream := KlineStreamPattern(symbolName, intervalValue, kline)
				applogger.Info("SyncHistoricalKline: Store kline(%s) into DB, StartT: %v", recordLabel, klineStream.Kline.StartTime)
				metaCol.Store("", klineStream)
				continue
			} else {
				// for those existed item, try to update
				if eventStored.Kline.IsFinal {
					applogger.Warn("SyncHistoricalKline: kline(%s) starttime: %v has finished synchron, discard incomming historical kline.", recordLabel, eventStored.Kline.StartTime)
					continue
				} else {
					// update:
					filter := bson.D{{"kline.starttime", kline.OpenTime}}
					update := bson.D{{"$set",
						bson.D{
							{"kline.open", kline.Open},
							{"kline.close", kline.Close},
							{"kline.high", kline.High},
							{"kline.low", kline.Low},
							{"kline.volume", kline.Volume},
							{"kline.tradenum", kline.NumberOfTrades},
							{"kline.isfinal", true},
							{"kline.quotevolume", kline.Volume},
							{"kline.activebuyvolume", kline.TakerBuyBaseAssetVolume},
							{"kline.activebuyquotevolume", kline.TakerBuyQuoteAssetVolume},
						}}}
					metaCol.Collection.UpdateOne(context.TODO(), filter, update)
					applogger.Info("SyncHistoricalKline: update the kline(%s) event: %v", recordLabel, binance_connector.PrettyPrint(kline))
				}
			}
		}

		// update the synced time, till the last kline's starttime.
		filter := bson.D{{"symbol", symbolName}}
		var syncedStartTime uint64
		if len(klines) == 0 {
			// no klines, maybe this project coin not yet published at this time window
			applogger.Info("SyncHistoricalKline: kline(%s) NOT published at this time window: %d to %d", recordLabel, timePair[0], timePair[1])
			syncedStartTime = timePair[0]
		} else {
			syncedStartTime = klines[len(klines)-1].OpenTime
		}
		update := bson.D{{"$set", bson.D{{"starttime", syncedStartTime}}}}
		opts := options.Update().SetUpsert(true)
		result, err := syncFlagCol.Collection.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			applogger.Warn("SyncHistoricalKline: Update kline(%s) Sync Flag failed: %v", recordLabel, err)
		} else {
			hsFlag.finished = true
			applogger.Info("SyncHistoricalKline: Update kline(%s) Sync Flag succeeded: %v", recordLabel, binance_connector.PrettyPrint(result))
		}
	}
}
