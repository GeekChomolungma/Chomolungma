package binance

import (
	binance_connector "github.com/binance/binance-connector-go"
)

func KlineStreamPattern(symbol, interval string, kline *binance_connector.KlinesResponse) *binance_connector.WsKlineEvent {
	klineStream := &binance_connector.WsKlineEvent{}
	klineStream.Time = int64(kline.CloseTime)
	klineStream.Symbol = symbol

	klineInner := binance_connector.WsKline{
		StartTime:            int64(kline.OpenTime),
		EndTime:              int64(kline.CloseTime),
		Symbol:               symbol,
		Interval:             interval,
		Open:                 kline.Open,
		Close:                kline.Close,
		High:                 kline.High,
		Low:                  kline.Low,
		Volume:               kline.Volume,
		TradeNum:             int64(kline.NumberOfTrades),
		IsFinal:              true,
		QuoteVolume:          kline.QuoteAssetVolume,
		ActiveBuyVolume:      kline.TakerBuyBaseAssetVolume,
		ActiveBuyQuoteVolume: kline.TakerBuyQuoteAssetVolume,
	}
	klineStream.Kline = klineInner
	return klineStream
}
