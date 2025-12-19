package internal

import (
	"time"
)

type CoinMarketCapResponse struct {
    Status CoinMarketCapResponseStatus `json:"status"`
    Data CoinMarketCapResponseDataEntry `json:"data"`
}

type CoinMarketCapResponseStatus struct {
    Timestamp time.Time `json:"timestamp"`
    ErrorCode int64 `json:"error_code"`
    ErrorMessage string `json:"error_message"`
    Elapsed int64 `json:"elapsed"`
    CreditCount int64 `json:"credit_count"`
    Notice string `json:"notice"`
}

type CoinMarketCapResponseDataEntry struct {
	BtcData CoinMarketCapResponseData `json:"1"`
}

type CoinMarketCapResponseData struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Symbol string `json:"symbol"`
	Slug string `json:"slug"`
	MaxSupply string `json:"max_supply"`
	LastUpdated time.Time `json:"last_updated:"`
	Quote CoinMarketCapCurrencyQuote `json:"quote"`
}

type CoinMarketCapCurrencyQuote struct {
	Usd CoinMarketCapQuote `json:"USD"`
}

type CoinMarketCapQuote struct {
	Price float64 `json:"price"`
	MarketCap float64 `json:"market_cap"`
	MarketCapDominance float64 `json:"market_cap_dominance"`
	LastUpdated time.Time `json:"last_updated"`
}