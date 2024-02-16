package priceservice

import (
	"encoding/json"
	"fmt"
	"math"
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

func (e *RestService) Query(params map[string]string) ([]types.SubmitPrice, error) {
	resp, err := grequests.Get(
		e.url,
		&grequests.RequestOptions{
			Params:         params,
			RequestTimeout: e.timeout,
		},
	)
	if err != nil {
		return []types.SubmitPrice{}, err
	}

	var priceData PriceData
	err = json.Unmarshal(resp.Bytes(), &priceData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return []types.SubmitPrice{}, err
	}

	// Convert PriceData to an array of SubmitPrice
	var submitPrices []types.SubmitPrice
	for symbol, price := range priceData.Prices {
		submitPrice := types.SubmitPrice{
			Symbol: symbol,
			Price:  uint64(price * math.Pow10(9)), // Assuming you want to convert the float64 price to uint64
		}
		submitPrices = append(submitPrices, submitPrice)
	}

	return submitPrices, nil
}
