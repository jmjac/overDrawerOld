package btcState

type Stats struct {
	Hour       Summary `json:"hour"`
	Day        Summary `json:"day"`
	Week       Summary `json:"week"`
	Month      Summary `json:"month"`
	Year       Summary `json:"year"`
	BlockCount int64   `json:"block_count"`
}

type Summary struct {
	TxCount int64   `json:"tx_count"`
	Value   float64 `json:"value"`
}
