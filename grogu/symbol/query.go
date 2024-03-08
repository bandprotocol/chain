package symbol

import (
	"math"
	"strconv"
	"strings"
	"time"

	bothanproto "github.com/bandprotocol/bothan-api/go-proxy/proto"
	"golang.org/x/exp/maps"

	"github.com/bandprotocol/chain/v2/grogu/grogucontext"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func StartQuerySymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		querySymbols(c, l)
	}
}

func querySymbols(c *grogucontext.Context, l *grogucontext.Logger) {
	symbolsWithTimeLimit := <-c.PendingSymbols

GetAllSymbols:
	for {
		select {
		case nextSymbols := <-c.PendingSymbols:
			maps.Copy(symbolsWithTimeLimit, nextSymbols)
		default:
			break GetAllSymbols
		}
	}

	symbols := maps.Keys(symbolsWithTimeLimit)

	l.Info("Try to get prices for symbols: %+v", symbols)
	prices, err := c.PriceService.Query(symbols)
	if err != nil {
		l.Error(":exploding_head: Failed to get prices from price-service with error: %s", c, err.Error())
	}

	maxSafePrice := math.MaxUint64 / uint64(math.Pow10(9))
	now := time.Now()
	submitPrices := []types.SubmitPrice{}
	for _, priceData := range prices {
		switch priceData.PriceOption {
		case bothanproto.PriceOption_PRICE_OPTION_UNSUPPORTED:
			submitPrices = append(submitPrices, types.SubmitPrice{
				PriceOption: types.PriceOptionUnsupported,
				Symbol:      priceData.SignalId,
				Price:       0,
			})
			continue
		case bothanproto.PriceOption_PRICE_OPTION_AVAILABLE:
			price, err := strconv.ParseFloat(strings.TrimSpace(priceData.Price), 64)
			if err != nil || price > float64(maxSafePrice) || price < 0 {
				l.Error(":exploding_head: Failed to parse price from symbol:", c, priceData.SignalId, err)
				priceData.PriceOption = bothanproto.PriceOption_PRICE_OPTION_UNAVAILABLE
				priceData.Price = ""
			} else {
				submitPrices = append(submitPrices, types.SubmitPrice{
					PriceOption: types.PriceOptionAvailable,
					Symbol:      priceData.SignalId,
					Price:       uint64(price * math.Pow10(9)),
				})
				continue
			}
		}

		if symbolsWithTimeLimit[priceData.SignalId].Before(now) {
			submitPrices = append(submitPrices, types.SubmitPrice{
				PriceOption: types.PriceOptionUnavailable,
				Symbol:      priceData.SignalId,
				Price:       0,
			})
		}
	}

	// delete symbol from in progress map if its price is not found
	symbolPriceMap := convertToSymbolPriceMap(submitPrices)
	for _, symbol := range symbols {
		if _, found := symbolPriceMap[symbol]; !found {
			c.InProgressSymbols.Delete(symbol)
		}
	}

	if len(submitPrices) == 0 {
		l.Debug(":exploding_head: query symbol got no prices with symbols: %+v", symbols)
		return
	}
	l.Info("got prices for symbols: %+v", maps.Keys(symbolPriceMap))
	c.PendingPrices <- submitPrices
}

// convertToSymbolPriceMap converts an array of SubmitPrice to a map of symbol to price.
func convertToSymbolPriceMap(data []types.SubmitPrice) map[string]uint64 {
	symbolPriceMap := make(map[string]uint64)

	for _, entry := range data {
		symbolPriceMap[entry.Symbol] = entry.Price
	}

	return symbolPriceMap
}
