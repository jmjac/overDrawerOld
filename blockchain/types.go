package blockchain

type Stats struct {
	Hour         Summary   `json:"hour"`
	Day          Summary   `json:"day"`
	Week         Summary   `json:"week"`
	Month        Summary   `json:"month"`
	MonthPerHour []Summary `json:"month_per_hour"`
	BlockCount   int       `json:"block_count"`
}

type Summary struct {
	TransactionCount int   `json:"transaction_count"`
	MoneyMoved       int64 `json:"money_moved"`
}
