package market

import (
	"github.com/huobirdcenter/huobi_golang/pkg/model/base"
	"github.com/shopspring/decimal"
)

type SubscribeCandlestickResponse struct {
	base base.WebSocketResponseBase
	Tick *Tick
	Data []Tick
}
type Tick struct {
	Id     int64           `json:"id"`
	Amount decimal.Decimal `json:"amount"`
	Count  int             `json:"count"`
	Open   decimal.Decimal `json:"open"`
	Close  decimal.Decimal `json:"close"`
	Low    decimal.Decimal `json:"low"`
	High   decimal.Decimal `json:"high"`
	Vol    decimal.Decimal `json:"vol"`
}

func (t *Tick) TickToFloat() *TickFloat {
	amount, _ := t.Amount.Float64()
	open, _ := t.Open.Float64()
	close, _ := t.Close.Float64()
	low, _ := t.Low.Float64()
	high, _ := t.High.Float64()
	vol, _ := t.Vol.Float64()
	return &TickFloat{
		Id:     t.Id,
		Amount: amount,
		Count:  t.Count,
		Open:   open,
		Close:  close,
		Low:    low,
		High:   high,
		Vol:    vol,
	}
}

type TickFloat struct {
	Id     int64   `json:"id"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Vol    float64 `json:"vol"`
}
