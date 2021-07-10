package dtos

type BaseReqModel struct {
	AimSite string `json:"aimsite"`
	Method  string `json:"method"`
	Url     string `json:"url"`
	Body    string `json:"body"`
}

type BaseRspModel struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}
