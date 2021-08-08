package dtos

// httpModel is a wrapper for http client,
// which controls the Chomolungma.
type HttpReqModel struct {
	AimSite   string `json:"aimsite"`
	AccountID string `json:"accountid"`
	Body      string `json:"body"`
}

type HttpRspModel struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}
