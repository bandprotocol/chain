package priceservice

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcprice "github.com/bandprotocol/bothan-api/go-proxy/proto"
)

type GRPCService struct {
	url     string
	timeout time.Duration
}

func NewGRPCService(url string, timeout time.Duration) *GRPCService {
	return &GRPCService{url: url, timeout: timeout}
}

func (gs *GRPCService) Query(symbols []string) ([]types.SubmitPrice, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(gs.url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return []types.SubmitPrice{}, err
	}
	defer conn.Close()

	// Create a client instance using the connection.
	client := grpcprice.NewQueryClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	response, err := client.Prices(ctx, &grpcprice.QueryPricesRequest{Symbols: symbols})
	if err != nil {
		return []types.SubmitPrice{}, err
	}

	maxSafePrice := math.MaxUint64 / uint64(math.Pow10(9))

	var submitPrices []types.SubmitPrice
	for _, priceData := range response.Prices {
		if priceData.PriceOption != grpcprice.PriceOption_PRICE_OPTION_AVAILABLE {
			submitPrice := types.SubmitPrice{
				Symbol: priceData.Symbol,
				Price:  0,
				Error:  priceData.PriceOption.String(),
			}
			submitPrices = append(submitPrices, submitPrice)
			continue
		}

		price, err := strconv.ParseFloat(strings.TrimSpace(priceData.Price), 64)
		if err != nil {
			submitPrice := types.SubmitPrice{
				Symbol: priceData.Symbol,
				Price:  0,
				Error:  err.Error(),
			}
			submitPrices = append(submitPrices, submitPrice)
			continue
		}

		if price > float64(maxSafePrice) || price < 0 {
			submitPrice := types.SubmitPrice{
				Symbol: priceData.Symbol,
				Price:  0,
				Error:  fmt.Sprintf("received price is out of range for symbol %s", priceData.Symbol),
			}
			submitPrices = append(submitPrices, submitPrice)
			continue
		}

		submitPrice := types.SubmitPrice{
			Symbol: priceData.Symbol,
			Price:  uint64(price * math.Pow10(9)),
		}
		submitPrices = append(submitPrices, submitPrice)
	}

	return submitPrices, nil
}
