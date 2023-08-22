package binance

import (
	"context"

	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	binance_connector "github.com/binance/binance-connector-go"
)

const (
	baseURL = "https://api.binance.com"
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

func SyncHistoricalKline(symbolName, intervalValue string, startTime, endTime uint64) {
	// repeated time between (n-1)th window's end and nth window's begin
	// to fix
	timeWindow := calcTimeWindow(periodUnit(intervalValue), startTime, endTime)

	client := binance_connector.NewClient("", "", baseURL)

	for k, timePair := range timeWindow {
		// Klines
		klines, err := client.NewKlinesService().
			Symbol(symbolName).Interval(intervalValue).StartTime(timePair[0]).EndTime(timePair[1]).Do(context.Background())
		if err != nil {
			applogger.Warn("GetKline failed once: timeWindow(500) index is %d, start:%d, to:%d", k, timePair[0], timePair[1])
			continue
		} else {
			applogger.Info("GetKline succeeded once: timeWindow(500) index is %d, start:%d, to:%d, kline count: %v", k, timePair[0], timePair[1], len(klines))
		}

		metaCol := mongoInc.NewMetaCollection[*binance_connector.WsKlineEvent]("marketInfo", symbolName, mongoInc.BinanKline)
		for _, kline := range klines {
			klineStream := KlineStreamPattern(symbolName, intervalValue, kline)
			applogger.Info("Store Kline into DB, StartT: %v", klineStream.Kline.StartTime)
			metaCol.Store("", klineStream)
		}
		//fmt.Println(binance_connector.PrettyPrint(klines))
	}
}
