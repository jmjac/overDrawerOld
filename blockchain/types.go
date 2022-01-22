package blockchain

type Stats struct {
	Hour        Summary   `json:"hour"`
	Day         Summary   `json:"day"`
	Week        Summary   `json:"week"`
	Month       Summary   `json:"month"`
	Year        Summary   `json:"year"`
	YearPerHour []Summary `json:"year_per_hour"`
	BlockCount  int       `json:"block_count"`
}

type Summary struct {
	TxCount int64 `json:"tx_count"`
	Value   int64 `json:"value"`
}
