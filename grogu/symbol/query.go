package symbol

import (
	"strings"

	"github.com/bandprotocol/chain/v2/grogu/grogucontext"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func StartQuerySymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		querySymbols(c, l)
	}
}

func querySymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	symbols := <-c.PendingSymbols

GetAllSymbols:
	for {
		select {
		case nextSymbols := <-c.PendingSymbols:
			symbols = append(symbols, nextSymbols...)
		default:
			break GetAllSymbols
		}
	}

	symbolStr := strings.Join(symbols, ",")

	params := map[string]string{
		"symbols": symbolStr,
	}

	l.Info("Try to get prices for symbols: %s", symbolStr)
	prices, err := c.PriceService.Query(params)
	if err != nil {
		l.Error(":exploding_head: Failed to get prices from price-service with error: %s", c, err.Error())
	}

	// delete symbol from in progress map if its price is not found
	symbolPriceMap := convertToSymbolPriceMap(prices)
	for _, symbol := range symbols {
		if _, found := symbolPriceMap[symbol]; !found {
			c.InProgressSymbols.Delete(symbol)
		}
	}

	l.Info("got prices for symbols: %s", symbolStr)
	if len(prices) == 0 {
		l.Error(":exploding_head: query symbol got no prices with symbols: %s", c, symbolStr)
		return
	}
	c.PendingPrices <- prices
}

// convertToSymbolPriceMap converts an array of SubmitPrice to a map of symbol to price.
func convertToSymbolPriceMap(data []types.SubmitPrice) map[string]uint64 {
	symbolPriceMap := make(map[string]uint64)

	for _, entry := range data {
		symbolPriceMap[entry.Symbol] = entry.Price
	}

	return symbolPriceMap
}
