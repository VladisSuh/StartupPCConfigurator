// internal/aggregator/usecase/messages.go
package usecase

// ShopUpdateMsg — тело сообщений из очереди shop_update
type ShopUpdateMsg struct {
	JobID  int64  `json:"jobId"`
	ShopID int64  `json:"shopId"`
	Type   string `json:"type,omitempty"`
}
