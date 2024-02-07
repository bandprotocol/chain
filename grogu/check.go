package grogu

import (
	"context"
	"time"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func checkSymbols(c *Context, l *Logger) {
	bz := cdc.MustMarshal(&types.QueryValidValidatorRequest{
		Validator: c.validator.String(),
	})
	resBz, err := c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/ValidValidator", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to check validator validity with error: %s", c, err.Error())
		return
	}

	validValidator := types.QueryValidValidatorResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &validValidator)
	if !validValidator.Valid {
		return
	}

	bz = cdc.MustMarshal(&types.QueryParamsRequest{})
	resBz, err = c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/Params", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get supported symbols with error: %s", c, err.Error())
		return
	}

	paramsResponse := types.QueryParamsResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &paramsResponse)
	params := paramsResponse.Params

	bz = cdc.MustMarshal(&types.QuerySupportedSymbolsRequest{})
	resBz, err = c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/SupportedSymbols", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get supported symbols with error: %s", c, err.Error())
		return
	}

	symbolsResponse := types.QuerySupportedSymbolsResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &symbolsResponse)
	symbols := symbolsResponse.Symbols

	var symbolList []string

	bz = cdc.MustMarshal(&types.QueryValidatorPricesRequest{
		Validator: c.validator.String(),
	})
	resBz, err = c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/ValidatorPrices", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get validator prices with error: %s", c, err.Error())
		return
	}
	validatorPricesResponse := types.QueryValidatorPricesResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &validatorPricesResponse)
	validatorPrices := validatorPricesResponse.ValidatorPrices
	symbolTimestampMap := ConvertToSymbolTimestampMap(validatorPrices)

	now := time.Now()

	for _, symbol := range symbols {
		if _, inProgress := c.inProgressSymbols.Load(symbol.GetSymbol()); inProgress {
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
			c.inProgressSymbols.Store(symbol.GetSymbol(), time.Now())
		}
	}
	if len(symbolList) != 0 {
		l.Info("found symbols to send: %v", symbolList)
		c.pendingSymbols <- symbolList
	}
}

func StartCheckSymbols(c *Context, l *Logger) {
	for {
		checkSymbols(c, l)
		time.Sleep(time.Second)
	}
}
