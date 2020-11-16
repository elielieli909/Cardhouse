package book

// OrderSchema defines the schema for an order received over http
type OrderSchema struct {
	Symbol      string `json:"symbol"`
	Qty         int    `json:"qty"`
	OrderType   string `json:"type"`
	Side        string `json:"side"`
	LimitPrice  int    `json:"limit"`
	TimeInForce string `json:"time_in_force"`
}
