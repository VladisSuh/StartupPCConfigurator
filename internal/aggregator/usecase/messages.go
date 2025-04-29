// internal/aggregator/usecase/messages.go
package usecase

// ShopUpdateMsg — тело сообщений из очереди shop_update
type ShopUpdateMsg struct {
	JobID  int64  `json:"jobId"`
	ShopID int64  `json:"shopId"`
	Type   string `json:"type,omitempty"`
}

type PriceChangedMsg struct {
	ComponentID string  `json:"componentId"`
	ShopID      int64   `json:"shopId"`
	OldPrice    float64 `json:"oldPrice"`
	NewPrice    float64 `json:"newPrice"`
	Timestamp   int64   `json:"timestamp"` // тут UnixNano или Unix
}
