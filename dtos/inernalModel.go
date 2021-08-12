package dtos

// HttpInternalKey for querying accountID
type HttpInternalAccountID struct {
	AimSite   string `json:"aimsite"`
	AccessKey string `json:"accesskey"`
	SecretKey string `json:"secretkey"`
}
