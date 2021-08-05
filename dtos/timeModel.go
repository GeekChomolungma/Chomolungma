package dtos

// SyncTimestampFlag is used to mark sync latest timestamp
type PreviousSyncTime struct {
	CollectionName string `json:"collectionname"`
	PreviousSyncTs int64  `json:"PreviousSyncTs"`
}
