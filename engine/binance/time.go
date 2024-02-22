package binance

import (
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
)

type periodUnit string

const (
	Period_1min  periodUnit = "1m"
	Period_5min  periodUnit = "5m"
	Period_15min periodUnit = "15m"
	Period_30min periodUnit = "30m"
	Period_60min periodUnit = "1h"
	Period_4hour periodUnit = "4h"
	Period_1day  periodUnit = "1d"
	Period_1week periodUnit = "1w"
	Period_1mon  periodUnit = "1m"
	Period_1year periodUnit = "1y"
)

func calcTimeWindow(interval periodUnit, startTs, endTs uint64) [][]uint64 {
	var divisor uint64
	var timeWindow [][]uint64
	switch interval {
	case Period_1min:
		divisor = 60 * 1000
	case Period_5min:
		divisor = 300 * 1000
	case Period_15min:
		divisor = 900 * 1000
	case Period_30min:
		divisor = 1800 * 1000
	case Period_60min:
		divisor = 3600 * 1000
	case Period_4hour:
		divisor = 14400 * 1000
	case Period_1day:
		divisor = 86400 * 1000
	case Period_1week:
		divisor = 604800 * 1000
	}

	newStartT := startTs
	windowLength := (endTs - startTs) / (divisor * 499)
	for i := 0; i < int(windowLength); i++ {
		timeWindow = append(timeWindow, []uint64{newStartT, newStartT + divisor*499})
		newStartT += divisor * 499
	}
	if newStartT < endTs {
		timeWindow = append(timeWindow, []uint64{newStartT, endTs})
	}

	applogger.Info("calcTimeWindow: timeWindow(500) length is %d, start:%d, to:%d", len(timeWindow), startTs, endTs)
	return timeWindow
}
