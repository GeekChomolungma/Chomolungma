package common

import "github.com/shopspring/decimal"

type GetSymbolsResponse struct {
	Status string   `json:"status"`
	Data   []Symbol `json:"data"`
}

type Symbol struct {
	BaseCurrency           string          `json:"base-currency"`
	QuoteCurrency          string          `json:"quote-currency"`
	PricePrecision         int             `json:"price-precision"`
	AmountPrecision        int             `json:"amount-precision"`
	SymbolPartition        string          `json:"symbol-partition"`
	Symbol                 string          `json:"symbol"`
	State                  string          `json:"state"`
	ValuePrecision         int             `json:"value-precision"`
	LimitOrderMinOrderAmt  decimal.Decimal `json:"limit-order-min-order-amt"`
	LimitOrderMaxOrderAmt  decimal.Decimal `json:"limit-order-max-order-amt"`
	SellMarketMinOrderAmt  decimal.Decimal `json:"sell-market-min-order-amt"`
	SellMarketMaxOrderAmt  decimal.Decimal `json:"sell-market-max-order-amt"`
	BuyMarketMaxOrderValue decimal.Decimal `json:"buy-market-max-order-value"`
	MinOrderValue          decimal.Decimal `json:"min-order-value"`
	MaxOrderValue          decimal.Decimal `json:"max-order-value"`
	LeverageRatio          decimal.Decimal `json:"leverage-ratio"`
}

func (s *Symbol) SymbolToFloat() *SymbolFloat {
	limitOrderMinOrderAmt, _ := s.LimitOrderMinOrderAmt.Float64()
	LimitOrderMaxOrderAmt, _ := s.LimitOrderMaxOrderAmt.Float64()
	sellMarketMinOrderAmt, _ := s.SellMarketMinOrderAmt.Float64()
	sellMarketMaxOrderAmt, _ := s.SellMarketMaxOrderAmt.Float64()
	buyMarketMaxOrderValue, _ := s.BuyMarketMaxOrderValue.Float64()
	minOrderValue, _ := s.MinOrderValue.Float64()
	maxOrderValue, _ := s.MaxOrderValue.Float64()
	leverageRatio, _ := s.LeverageRatio.Float64()

	return &SymbolFloat{
		BaseCurrency:           s.BaseCurrency,
		QuoteCurrency:          s.QuoteCurrency,
		PricePrecision:         s.PricePrecision,
		AmountPrecision:        s.AmountPrecision,
		SymbolPartition:        s.SymbolPartition,
		Symbol:                 s.Symbol,
		State:                  s.State,
		ValuePrecision:         s.ValuePrecision,
		LimitOrderMinOrderAmt:  limitOrderMinOrderAmt,
		LimitOrderMaxOrderAmt:  LimitOrderMaxOrderAmt,
		SellMarketMinOrderAmt:  sellMarketMinOrderAmt,
		SellMarketMaxOrderAmt:  sellMarketMaxOrderAmt,
		BuyMarketMaxOrderValue: buyMarketMaxOrderValue,
		MinOrderValue:          minOrderValue,
		MaxOrderValue:          maxOrderValue,
		LeverageRatio:          leverageRatio,
	}
}

// SymbolFloat ...
type SymbolFloat struct {
	BaseCurrency           string  `json:"base-currency"`
	QuoteCurrency          string  `json:"quote-currency"`
	PricePrecision         int     `json:"price-precision"`
	AmountPrecision        int     `json:"amount-precision"`
	SymbolPartition        string  `json:"symbol-partition"`
	Symbol                 string  `json:"symbol"`
	State                  string  `json:"state"`
	ValuePrecision         int     `json:"value-precision"`
	LimitOrderMinOrderAmt  float64 `json:"limit-order-min-order-amt"`
	LimitOrderMaxOrderAmt  float64 `json:"limit-order-max-order-amt"`
	SellMarketMinOrderAmt  float64 `json:"sell-market-min-order-amt"`
	SellMarketMaxOrderAmt  float64 `json:"sell-market-max-order-amt"`
	BuyMarketMaxOrderValue float64 `json:"buy-market-max-order-value"`
	MinOrderValue          float64 `json:"min-order-value"`
	MaxOrderValue          float64 `json:"max-order-value"`
	LeverageRatio          float64 `json:"leverage-ratio"`
}
