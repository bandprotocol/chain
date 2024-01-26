package grogu

import "strings"

func StartQuerySymbols(c *Context, l *Logger) {
	for {
		querySymbols(c, l)
	}
}

func querySymbols(c *Context, l *Logger) {
	symbols := <-c.pendingSymbols

GetAllSymbols:
	for {
		select {
		case nextSymbols := <-c.pendingSymbols:
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
	prices, err := c.priceService.Query(params)
	if err != nil {
		l.Error(":exploding_head: Failed to get prices from price-service with error: %s", c, err.Error())
	}

	// delete symbol from in progress map if its price is not found
	symbolPriceMap := ConvertToSymbolPriceMap(prices)
	for _, symbol := range symbols {
		if _, found := symbolPriceMap[symbol]; !found {
			c.inProgressSymbols.Delete(symbol)
		}
	}

	l.Info("got prices for symbols: %s", symbolStr)
	if len(prices) == 0 {
		l.Error(":exploding_head: query symbol got no prices with symbols: %s", c, symbolStr)
		return
	}
	c.pendingPrices <- prices
}
