package symbol

import (
	"context"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/bandprotocol/chain/v2/grogu/grogucontext"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func checkSymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	bz := grogucontext.Cdc.MustMarshal(&types.QueryValidValidatorRequest{
		Validator: c.Validator.String(),
	})
	resBz, err := abciQuery(c, l, "/feeds.v1beta1.Query/ValidValidator", bz)
	if err != nil {
		return
	}

	validValidator := types.QueryValidValidatorResponse{}
	grogucontext.Cdc.MustUnmarshal(resBz.Response.Value, &validValidator)
	if !validValidator.Valid {
		return
	}

	bz = grogucontext.Cdc.MustMarshal(&types.QueryParamsRequest{})
	resBz, err = abciQuery(c, l, "/feeds.v1beta1.Query/Params", bz)
	if err != nil {
		return
	}

	paramsResponse := types.QueryParamsResponse{}
	grogucontext.Cdc.MustUnmarshal(resBz.Response.Value, &paramsResponse)
	params := paramsResponse.Params

	bz = grogucontext.Cdc.MustMarshal(&types.QuerySupportedSymbolsRequest{})
	resBz, err = abciQuery(c, l, "/feeds.v1beta1.Query/SupportedSymbols", bz)
	if err != nil {
		return
	}

	symbolsResponse := types.QuerySupportedSymbolsResponse{}
	grogucontext.Cdc.MustUnmarshal(resBz.Response.Value, &symbolsResponse)
	symbols := symbolsResponse.Symbols

	var symbolList []string

	bz = grogucontext.Cdc.MustMarshal(&types.QueryValidatorPricesRequest{
		Validator: c.Validator.String(),
	})
	resBz, err = abciQuery(c, l, "/feeds.v1beta1.Query/ValidatorPrices", bz)
	if err != nil {
		return
	}
	validatorPricesResponse := types.QueryValidatorPricesResponse{}
	grogucontext.Cdc.MustUnmarshal(resBz.Response.Value, &validatorPricesResponse)
	validatorPrices := validatorPricesResponse.ValidatorPrices
	symbolTimestampMap := convertToSymbolTimestampMap(validatorPrices)

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
			symbolList = append(symbolList, symbol.Symbol)
			c.InProgressSymbols.Store(symbol.GetSymbol(), time.Now())
		}
	}
	if len(symbolList) != 0 {
		l.Info("found symbols to send: %v", symbolList)
		c.PendingSymbols <- symbolList
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

// abciQuery will try to query data from BandChain node
func abciQuery(
	c *grogucontext.Context,
	l *grogucontext.Logger,
	path string,
	data []byte,
) (*ctypes.ResultABCIQuery, error) {
	var lastErr error
	res, err := c.Client.ABCIQuery(context.Background(), path, data)
	if err != nil {
		l.Debug(":exploding_head: Failed to query on %s request with error: %s", path, err.Error())
		return nil, lastErr
	}
	return res, nil
}

func StartCheckSymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		checkSymbols(c, l)
		time.Sleep(time.Second)
	}
}
