package trading

import (
	"fmt"
	"math/rand"
	"time"
)

type TickerData struct {
	Symbol string
	Price  string
}

type Engine struct {
	rng *rand.Rand
}

func NewEngine() *Engine {
	return &Engine{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// FetchTicker simulates an API call without external SDKs
func (e *Engine) FetchTicker(symbol string) (*TickerData, error) {
	// Simulate network latency
	time.Sleep(10 * time.Millisecond)

	// Simulate price fluctuation
	basePrice := 65000.0
	variation := (e.rng.Float64() * 1000) - 500
	price := basePrice + variation

	return &TickerData{
		Symbol: symbol,
		Price:  fmt.Sprintf("%.2f", price),
	}, nil
}
