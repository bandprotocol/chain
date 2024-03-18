package symbol

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"golang.org/x/exp/maps"

	band "github.com/bandprotocol/chain/v2/app"
	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func checkSymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	clientCtx := client.Context{
		Client:            c.Client,
		Codec:             grogucontext.Cdc,
		TxConfig:          band.MakeEncodingConfig().TxConfig,
		BroadcastMode:     flags.BroadcastSync,
		InterfaceRegistry: band.MakeEncodingConfig().InterfaceRegistry,
	}

	queryClient := types.NewQueryClient(clientCtx)

	validValidator, err := queryClient.ValidValidator(context.Background(), &types.QueryValidValidatorRequest{
		Validator: c.Validator.String(),
	})
	if err != nil {
		return
	}

	if !validValidator.Valid {
		return
	}

	paramsResponse, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return
	}
	params := paramsResponse.Params

	symbolsResponse, err := queryClient.SupportedSymbols(context.Background(), &types.QuerySupportedSymbolsRequest{})
	if err != nil {
		return
	}

	symbols := symbolsResponse.Symbols

	validatorPricesResponse, err := queryClient.ValidatorPrices(
		context.Background(),
		&types.QueryValidatorPricesRequest{
			Validator: c.Validator.String(),
		},
	)
	if err != nil {
		return
	}

	validatorPrices := validatorPricesResponse.ValidatorPrices
	symbolTimestampMap := convertToSymbolTimestampMap(validatorPrices)

	requestedSymbols := make(map[string]time.Time)
	now := time.Now()

	for _, symbol := range symbols {
		if _, inProgress := c.InProgressSymbols.Load(symbol.GetSymbol()); inProgress {
			continue
		}

		timestamp, ok := symbolTimestampMap[symbol.GetSymbol()]
		// add 2 to prevent too fast cases
		if !ok ||
			time.Unix(timestamp+2, 0).
				Add(time.Duration(symbol.Interval)*time.Second).
				Add(-time.Duration(params.TransitionTime)*time.Second).
				Before(now) {
			requestedSymbols[symbol.Symbol] = time.Unix(timestamp, 0).
				Add(time.Duration(symbol.Interval) * time.Second).
				Add(-time.Duration(params.TransitionTime) * time.Second / 2)
			c.InProgressSymbols.Store(symbol.GetSymbol(), time.Now())
		}
	}
	if len(requestedSymbols) != 0 {
		l.Info("found symbols to send: %v", maps.Keys(requestedSymbols))
		c.PendingSymbols <- requestedSymbols
	}
}

// convertToSymbolTimestampMap converts an array of PriceValidator to a map of symbol to timestamp.
func convertToSymbolTimestampMap(data []types.PriceValidator) map[string]int64 {
	symbolTimestampMap := make(map[string]int64)

	for _, entry := range data {
		symbolTimestampMap[entry.Symbol] = entry.Timestamp
	}

	return symbolTimestampMap
}

func StartCheckSymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		checkSymbols(c, l)
		time.Sleep(time.Second)
	}
}
