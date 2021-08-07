package dtos

// OrderPlace Parse bizniz from Climber
type OrderPlace struct {
	Symbol string `json:"symbol"` // like "btcusdt"
	Model  string `json:"model"`  // buy-market, sell-market, buy-limit, sell-limit...
	Amount string `json:"amount"` // order vol(order amount for market model)
	Price  string `json:"price"`  // useless for market model
	Source string `json:"source"` // spot-api, margin-api, super-margin-api, c2c-margin-api
}

// OrderCancel parse cancel req from chomo
type OrderCancel struct {
	OrderID string `json:"orderid"`
}

type OrderQuery struct {
	OrderID string `json:"orderid"`
}

// CurrencyBalanceReq request the currency's balance
type CurrencyBalanceReq struct {
	Currency string `json:"currency"`
}
