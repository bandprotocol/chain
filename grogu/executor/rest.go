package executor

import (
	"encoding/json"
	"fmt"
	"time"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/levigross/grequests"
)

type RestExec struct {
	url     string
	timeout time.Duration
}

func NewRestExec(url string, timeout time.Duration) *RestExec {
	return &RestExec{url: url, timeout: timeout}
}

type externalExecutionResponse struct {
	Returncode uint32 `json:"returncode"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Version    string `json:"version"`
}

type PriceData struct {
	Prices map[string]float64 `json:"prices"`
}

func (e *RestExec) Exec(params map[string]string) ([]feedstypes.SubmitPrice, error) {
	fmt.Println("executor url", e.url)
	resp, err := grequests.Get(
		e.url,
		&grequests.RequestOptions{
			Params: params,
		},
	)

	if err != nil {
		return []feedstypes.SubmitPrice{}, err
	}

	var priceData PriceData
	err = json.Unmarshal(resp.Bytes(), &priceData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return []feedstypes.SubmitPrice{}, err
	}

	// Convert PriceData to an array of SubmitPrice
	var submitPrices []feedstypes.SubmitPrice
	for symbol, price := range priceData.Prices {
		submitPrice := feedstypes.SubmitPrice{
			Symbol: symbol,
			Price:  uint64(price), // Assuming you want to convert the float64 price to uint64
		}
		submitPrices = append(submitPrices, submitPrice)
	}

	return submitPrices, nil
}
