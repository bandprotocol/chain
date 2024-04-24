package feed_test

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

	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/grogu/feed"
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
func (mps *MockPriceService) Query(signalIDs []string) ([]*bothanproto.PriceData, error) {
	var priceData []*bothanproto.PriceData
	for _, id := range signalIDs {
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

func TestQuerySignalIDs(t *testing.T) {
	// Create a mock context and logger for testing.
	mockContext := grogucontext.Context{}
	mockLogger := grogucontext.NewLogger(log.AllowAll())

	// Mock the pending signalIDs channel.
	mockContext.PendingSignalIDs = make(chan map[string]time.Time, 10)
	mockContext.PendingPrices = make(chan []types.SubmitPrice, 10)

	// Test cases: price available and price not supported
	signalIDsWithTimeLimit := make(map[string]time.Time)
	mockContext.InProgressSignalIDs = &sync.Map{}

	signalIDsWithTimeLimit["BTC"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSignalIDs.Load("BTC")

	signalIDsWithTimeLimit["DOGE"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSignalIDs.Load("DOGE")

	mockContext.PendingSignalIDs <- signalIDsWithTimeLimit

	// Set up a mock price service.
	mockContext.PriceService = &MockPriceService{}

	// Call the function being tested.
	feed.QuerySignalIDs(&mockContext, mockLogger)

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
	mockContext.InProgressSignalIDs.Delete("BTC")
	mockContext.InProgressSignalIDs.Delete("DOGE")

	signalIDsWithTimeLimit = make(map[string]time.Time)
	signalIDsWithTimeLimit["ETH"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSignalIDs.Load("ETH")
	signalIDsWithTimeLimit["BAND"] = time.Now().
		Add(time.Minute)
	mockContext.InProgressSignalIDs.Load("BAND")
	mockContext.PendingSignalIDs <- signalIDsWithTimeLimit

	// Call the function being tested.
	feed.QuerySignalIDs(&mockContext, mockLogger)

	// Check if the correct prices were sent to the pending prices channel.
	select {
	case submitPrices := <-mockContext.PendingPrices:
		t.Error("Should receive no prices but receive:", submitPrices)
	default:
	}

	// Test cases: price out of range, price unavailable with time limit reached
	signalIDsWithTimeLimit = make(map[string]time.Time)
	signalIDsWithTimeLimit["ETH"] = time.Now().
		Add(-time.Minute)
	mockContext.InProgressSignalIDs.Load("ETH")
	signalIDsWithTimeLimit["BAND"] = time.Now().
		Add(-time.Minute)
	mockContext.InProgressSignalIDs.Load("BAND")
	mockContext.PendingSignalIDs <- signalIDsWithTimeLimit

	// Call the function being tested.
	feed.QuerySignalIDs(&mockContext, mockLogger)

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

// getPrice retrieves the price data for a specific signalID from the submit prices array.
func getPrice(submitPrices []types.SubmitPrice, signalID string) types.SubmitPrice {
	for _, price := range submitPrices {
		if price.SignalID == signalID {
			return price
		}
	}
	return types.SubmitPrice{}
}
