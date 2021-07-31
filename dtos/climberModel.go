package dtos

// OrderInfo Parse bizniz from Climber
type OrderInfo struct {
	Model  string `json:"model"`  // buy-market, sell-market, buy-limit, sell-limit...
	Amount string `json:"amount"` // order vol(order amount for market model)
	Price  string `json:"price"`  // useless for market model
}