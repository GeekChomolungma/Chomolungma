package dtos

// SyncTimestampFlag is used to mark sync latest timestamp
type PreviousSyncTime struct {
	CollectionName string `json:"collectionname"`
	LatestSyncTs   int64  `json:"latestsyncts"`
}
