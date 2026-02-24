package trading

import (
	"encoding/json"
	"fmt"

	v2 "github.com/s4mn0v/bitget/pkg/client/v2"
)

type TickerData struct {
	Symbol string `json:"symbol"`
	Price  string `json:"lastPr"`
}

type tickerResponse struct {
	Data []TickerData `json:"data"`
}

type Engine struct {
	market *v2.MixMarketClient
}

func NewEngine() *Engine {
	return &Engine{market: new(v2.MixMarketClient).Init()}
}

// Method name must be FetchTicker
func (e *Engine) FetchTicker(symbol string) (*TickerData, error) {
	params := map[string]string{
		"symbol":      symbol,
		"productType": "USDT-FUTURES",
	}

	resp, err := e.market.Ticker(params)
	if err != nil {
		return nil, err
	}

	var tr tickerResponse
	if err := json.Unmarshal([]byte(resp), &tr); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if len(tr.Data) > 0 {
		return &tr.Data[0], nil
	}

	return nil, fmt.Errorf("no data for %s", symbol)
}
