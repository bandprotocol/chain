package symbol_test

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	bothanproto "github.com/bandprotocol/bothan-api/go-proxy/proto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/grogu/grogucontext"
	"github.com/bandprotocol/chain/v2/grogu/symbol"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

var mockData = map[string]*bothanproto.PriceData{
	"BTC":  {SignalId: "BTC", PriceOption: bothanproto.PriceOption_PRICE_OPTION_AVAILABLE, Price: "50000"},
	"ETH":  {SignalId: "ETH", PriceOption: bothanproto.PriceOption_PRICE_OPTION_UNAVAILABLE, Price: ""},
	"BAND": {SignalId: "BAND", PriceOption: bothanproto.PriceOption_PRICE_OPTION_AVAILABLE, Price: "18446744074"},
}

// MockPriceService is a mock implementation of the price service for testing.
type MockPriceService struct{}

// Query is a mock implementation of the Query method in the PriceService interface.
func (mps *MockPriceService) Query(signalIds []string) ([]*bothanproto.PriceData, error) {
	var priceData []*bothanproto.PriceData
	for _, id := range signalIds {
		data, ok := mockData[id]
		if ok {
			priceData = append(priceData, data)
		} else {
			priceData = append(priceData, &bothanproto.PriceData{
				SignalId:    id,
				PriceOption: bothanproto.PriceOption_PRICE_OPTION_UNSUPPORTED,
			})
		}
	}

	return priceData, nil
}

func TestQuerySymbols(t *testing.T) {
	// Create a mock context and logger for testing.
	mockContext := grogucontext.Context{}
	mockLogger := grogucontext.NewLogger(log.AllowAll())

	// Mock the pending symbols channel.
	mockContext.PendingSymbols = make(chan map[string]time.Time, 10)
	mockContext.PendingPrices = make(chan []types.SubmitPrice, 10)

	// Test cases: price available and price not supported
	symbolsWithTimeLimit := make(map[string]time.Time)
	mockContext.InProgressSymbols = &sync.Map{}

	symbolsWithTimeLimit["BTC"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSymbols.Load("BTC")

	symbolsWithTimeLimit["DOGE"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSymbols.Load("DOGE")

	mockContext.PendingSymbols <- symbolsWithTimeLimit

	// Set up a mock price service.
	mockContext.PriceService = &MockPriceService{}

	// Call the function being tested.
	symbol.QuerySymbols(&mockContext, mockLogger)

	// Check if the correct prices were sent to the pending prices channel.
	select {
	case submitPrices := <-mockContext.PendingPrices:
		// Check the number of prices received.
		require.Equal(t, 2, len(submitPrices))

		// Check if BTC price is correct.
		btcPrice := getPrice(submitPrices, "BTC")
		require.Equal(t, types.PriceOptionAvailable, btcPrice.PriceOption)
		mockBTCPrice, _ := strconv.ParseFloat(strings.TrimSpace(mockData["BTC"].Price), 64)
		require.Equal(t, uint64(mockBTCPrice*math.Pow10(9)), btcPrice.Price)

		dogePrice := getPrice(submitPrices, "DOGE")
		require.Equal(t, types.PriceOptionUnsupported, dogePrice.PriceOption)
		require.Equal(t, uint64(0), dogePrice.Price)

	default:
		t.Error("No prices received")
	}

	// Test cases: price out of range, price unavailable with time limit not reached
	mockContext.InProgressSymbols.Delete("BTC")
	mockContext.InProgressSymbols.Delete("DOGE")

	symbolsWithTimeLimit = make(map[string]time.Time)
	symbolsWithTimeLimit["ETH"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSymbols.Load("ETH")
	symbolsWithTimeLimit["BAND"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSymbols.Load("BAND")
	mockContext.PendingSymbols <- symbolsWithTimeLimit

	// Call the function being tested.
	symbol.QuerySymbols(&mockContext, mockLogger)

	// Check if the correct prices were sent to the pending prices channel.
	select {
	case submitPrices := <-mockContext.PendingPrices:
		t.Error("Should receive no prices but receive:", submitPrices)
	default:
	}

	// Test cases: price out of range, price unavailable with time limit reached
	symbolsWithTimeLimit = make(map[string]time.Time)
	symbolsWithTimeLimit["ETH"] = time.Now().
		Add(-time.Minute)
	mockContext.InProgressSymbols.Load("ETH")
	symbolsWithTimeLimit["BAND"] = time.Now().
		Add(-time.Minute)
	mockContext.InProgressSymbols.Load("BAND")
	mockContext.PendingSymbols <- symbolsWithTimeLimit

	// Call the function being tested.
	symbol.QuerySymbols(&mockContext, mockLogger)

	// Check if the correct prices were sent to the pending prices channel.
	select {
	case submitPrices := <-mockContext.PendingPrices:
		// Check the number of prices received.
		require.Equal(t, 2, len(submitPrices))

		ethPrice := getPrice(submitPrices, "ETH")
		require.Equal(t, types.PriceOptionUnavailable, ethPrice.PriceOption)
		require.Equal(t, uint64(0), ethPrice.Price)

		bandPrice := getPrice(submitPrices, "BAND")
		require.Equal(t, types.PriceOptionUnavailable, bandPrice.PriceOption)
		require.Equal(t, uint64(0), bandPrice.Price)
	default:
		t.Error("No prices received")
	}
}

// getPrice retrieves the price data for a specific symbol from the submit prices array.
func getPrice(submitPrices []types.SubmitPrice, symbol string) types.SubmitPrice {
	for _, price := range submitPrices {
		if price.Symbol == symbol {
			return price
		}
	}
	return types.SubmitPrice{}
}
