package dtos

// HttpInternalKey for querying accountID
type HttpInternalAccountID struct {
	AimSite   string `json:"aimsite"`
	AccessKey string `json:"accesskey"`
	SecretKey string `json:"secretkey"`
}

type DBValidationReq struct {
	AimSite    string `json:"aimsite"`
	Secret     string `json:"secret"`
	Collection string `json:"collection"`
}
