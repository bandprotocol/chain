package priceservice

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/levigross/grequests"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

type RestService struct {
	url     string
	timeout time.Duration
}

func NewRestService(url string, timeout time.Duration) *RestService {
	return &RestService{url: url, timeout: timeout}
}

type PriceData struct {
	Prices map[string]float64 `json:"prices"`
}

func (rs *RestService) Query(symbols []string) ([]types.SubmitPrice, error) {
	symbolStr := strings.Join(symbols, ",")

	params := map[string]string{
		"symbols": symbolStr,
	}

	resp, err := grequests.Get(
		rs.url,
		&grequests.RequestOptions{
			Params:         params,
			RequestTimeout: rs.timeout,
		},
	)
	if err != nil {
		return []types.SubmitPrice{}, err
	}

	var priceData PriceData
	err = json.Unmarshal(resp.Bytes(), &priceData)
	if err != nil {
		return []types.SubmitPrice{}, err
	}

	maxSafePrice := math.MaxUint64 / uint64(math.Pow10(9))

	// Convert PriceData to an array of SubmitPrice
	var submitPrices []types.SubmitPrice
	for symbol, price := range priceData.Prices {
		if price > float64(maxSafePrice) || price < 0 {
			return []types.SubmitPrice{}, fmt.Errorf("received price is out of range for symbol %s", symbol)
		}
		submitPrice := types.SubmitPrice{
			Symbol: symbol,
			Price:  uint64(price * math.Pow10(9)), // Assuming you want to convert the float64 price to uint64
		}
		submitPrices = append(submitPrices, submitPrice)
	}

	return submitPrices, nil
}
