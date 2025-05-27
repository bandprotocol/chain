package signaller

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client/proto/bothan/v1"

	"github.com/bandprotocol/chain/v3/grogu/signaller/testutil"
	"github.com/bandprotocol/chain/v3/grogu/submitter"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	feeds "github.com/bandprotocol/chain/v3/x/feeds/types"
)

type SignallerTestSuite struct {
	suite.Suite

	Signaller    *Signaller
	SubmitCh     chan submitter.SignalPriceSubmission
	assignedTime time.Time
}

func TestSignallerTestSuite(t *testing.T) {
	suite.Run(t, new(SignallerTestSuite))
}

func (s *SignallerTestSuite) SetupTest() {
	// Set up validator address
	validAddress := sdk.ValAddress("1000000001")

	ctrl := gomock.NewController(s.T())
	mockFeedQuerier := testutil.NewMockFeedQuerier(ctrl)
	mockFeedQuerier.EXPECT().
		QueryValidValidator(gomock.Any()).
		Return(&feeds.QueryValidValidatorResponse{Valid: true}, nil).
		AnyTimes()
	mockFeedQuerier.EXPECT().
		QueryValidatorPrices(gomock.Any()).
		Return(&feeds.QueryValidatorPricesResponse{ValidatorPrices: []feeds.ValidatorPrice{
			{
				SignalPriceStatus: feeds.SIGNAL_PRICE_STATUS_AVAILABLE,
				SignalID:          "signal1",
				Price:             10000,
				Timestamp:         0,
			},
		}}, nil).
		AnyTimes()
	mockFeedQuerier.EXPECT().
		QueryParams().
		Return(&feeds.QueryParamsResponse{Params: feeds.DefaultParams()}, nil).
		AnyTimes()
	mockFeedQuerier.EXPECT().
		QueryCurrentFeeds().
		Return(&feeds.QueryCurrentFeedsResponse{CurrentFeeds: feeds.CurrentFeedWithDeviations{
			Feeds: []feeds.FeedWithDeviation{
				{
					SignalID:            "signal1",
					Power:               60000000000,
					Interval:            60,
					DeviationBasisPoint: 50,
				},
				{
					SignalID:            "signal2",
					Power:               60000000000,
					Interval:            60,
					DeviationBasisPoint: 50,
				},
			},
		}}, nil).
		AnyTimes()

	mockBothanClient := testutil.NewMockBothanClient(ctrl)
	mockBothanClient.EXPECT().GetPrices(gomock.Any()).
		Return(&bothan.GetPricesResponse{
			Prices: []*bothan.Price{
				{
					SignalId: "signal1",
					Price:    10000,
					Status:   bothan.Status_STATUS_AVAILABLE,
				},
			},
			Uuid: "uuid1",
		}, nil).
		AnyTimes()

	mockNodeQuerier := testutil.NewMockNodeQuerier(ctrl)
	mockNodeQuerier.EXPECT().QueryStatus().
		DoAndReturn(func() (*node.StatusResponse, error) {
			time := time.Unix(0, 0)
			return &node.StatusResponse{
				Timestamp: &time,
			}, nil
		}).
		AnyTimes()

	// Create submit channel
	submitCh := make(chan submitter.SignalPriceSubmission, 300)

	// Initialize logger
	allowLevel, _ := log.ParseLogLevel("info")
	l := logger.NewLogger(allowLevel)

	// Initialize pending signal IDs map
	pendingSignalIDs := sync.Map{}

	// Create signaller instance
	s.Signaller = New(
		mockFeedQuerier,
		mockNodeQuerier,
		mockBothanClient,
		time.Second,
		submitCh,
		l,
		validAddress,
		&pendingSignalIDs,
		50,
		30,
	)
	s.SubmitCh = submitCh
	s.assignedTime = calculateAssignedTime(
		s.Signaller.valAddress,
		60,
		0,
		s.Signaller.distributionStartPercentage,
		s.Signaller.distributionOffsetPercentage,
	)
}

func (s *SignallerTestSuite) TestUpdateInternalVariables() {
	success := s.Signaller.updateInternalVariables()
	s.Require().True(success)
	s.Require().NotNil(s.Signaller.params)
	s.Require().NotEmpty(s.Signaller.signalIDToFeed)
	s.Require().NotEmpty(s.Signaller.signalIDToValidatorPrice)
}

func (s *SignallerTestSuite) TestFilterAndPrepareSignalPrices() {
	s.TestUpdateInternalVariables()

	// Test with available price
	prices := []*bothan.Price{
		{
			SignalId: "signal1",
			Price:    10000,
			Status:   bothan.Status_STATUS_AVAILABLE,
		},
	}

	signalIDs := []string{"signal1"}
	// Test with time in the middle of the interval
	s.Signaller.currentBlockTime = time.Unix(60/2, 0)

	submitPrices := s.Signaller.filterAndPrepareSignalPrices(prices, signalIDs)
	s.Require().Empty(submitPrices)

	// Test with time at the end of the interval
	s.Signaller.currentBlockTime = time.Unix(60, 0)

	submitPrices = s.Signaller.filterAndPrepareSignalPrices(prices, signalIDs)
	s.Require().NotEmpty(submitPrices)
	s.Require().Equal("signal1", submitPrices[0].SignalID)
	s.Require().Equal(uint64(10000), submitPrices[0].Price)

	// Test with unavailable price
	prices = []*bothan.Price{
		{
			SignalId: "signal1",
			Price:    10000,
			Status:   bothan.Status_STATUS_UNAVAILABLE,
		},
	}

	// Test with time after the urgent deadline
	s.Signaller.currentBlockTime = time.Unix(60-FixedIntervalOffset+1, 0)
	submitPrices = s.Signaller.filterAndPrepareSignalPrices(prices, signalIDs)
	s.Require().NotEmpty(submitPrices)
	s.Require().Equal("signal1", submitPrices[0].SignalID)
	s.Require().Equal(uint64(0), submitPrices[0].Price)

	// Test with time before the urgent deadline
	s.Signaller.currentBlockTime = time.Unix(60-FixedIntervalOffset-1, 0)
	submitPrices = s.Signaller.filterAndPrepareSignalPrices(prices, signalIDs)
	s.Require().Empty(submitPrices)
}

func (s *SignallerTestSuite) TestGetAllSignalIDs() {
	signalIDs := s.Signaller.getAllSignalIDs()
	s.Require().Empty(signalIDs)

	// Update internal variables
	s.TestUpdateInternalVariables()

	expectedSignalIDs := []string{"signal1", "signal2"}

	signalIDs = s.Signaller.getAllSignalIDs()
	s.Require().NotEmpty(signalIDs)

	// sort signalIDs to compare
	sort.Strings(signalIDs)
	s.Require().Equal(expectedSignalIDs, signalIDs)
}

func (s *SignallerTestSuite) TestGetNonPendingSignalIDs() {
	signalIDs := s.Signaller.getNonPendingSignalIDs()
	s.Require().Empty(signalIDs)

	// Update internal variables
	s.TestUpdateInternalVariables()

	expectedSignalIDs := []string{"signal1", "signal2"}

	signalIDs = s.Signaller.getNonPendingSignalIDs()
	s.Require().NotEmpty(signalIDs)

	// sort signalIDs to compare
	sort.Strings(signalIDs)
	s.Require().Equal(expectedSignalIDs, signalIDs)
}

func (s *SignallerTestSuite) TestSignalPrices() {
	prices := []feeds.SignalPrice{
		{
			SignalID: "signal1",
			Price:    10000,
			Status:   feeds.SIGNAL_PRICE_STATUS_AVAILABLE,
		},
	}

	uuid := "test-uuid"

	s.Signaller.submitPrices(prices, uuid)

	select {
	case priceSubmission := <-s.SubmitCh:
		s.Require().NotEmpty(priceSubmission.SignalPrices)
		s.Require().Equal("signal1", priceSubmission.SignalPrices[0].SignalID)
	default:
		s.Fail("Expected prices to be submitted")
	}
}

func (s *SignallerTestSuite) TestIsPriceValid() {
	// Update internal variables
	s.TestUpdateInternalVariables()

	priceData := feeds.SignalPrice{
		Status: feeds.SIGNAL_PRICE_STATUS_AVAILABLE,
		Price:  10000,
	}

	// Test with price is not required to be submitted
	priceData.SignalID = "signal3"
	s.Signaller.currentBlockTime = s.assignedTime
	s.Require().False(s.Signaller.isPriceValid(priceData))

	// Test with price is required to be submitted and not exist yet
	priceData.SignalID = "signal2"
	s.Signaller.currentBlockTime = s.assignedTime
	s.Require().True(s.Signaller.isPriceValid(priceData))
}

func (s *SignallerTestSuite) TestShouldUpdatePrice() {
	// Update internal variables
	s.TestUpdateInternalVariables()

	feed := feeds.FeedWithDeviation{
		SignalID:            "signal1",
		Interval:            60,
		DeviationBasisPoint: 50,
	}

	valPrice := feeds.ValidatorPrice{
		SignalID:          "signal1",
		Price:             10000,
		Timestamp:         0,
		SignalPriceStatus: feeds.SIGNAL_PRICE_STATUS_AVAILABLE,
	}

	newPrice := feeds.SignalPrice{
		Price:  10000,
		Status: feeds.SIGNAL_PRICE_STATUS_AVAILABLE,
	}

	thresholdTime := time.Unix(valPrice.Timestamp+s.Signaller.params.CooldownTime, 0)

	// Test case: Time before thresholdTime, should not update
	s.Signaller.currentBlockTime = thresholdTime.Add(-time.Second)
	s.Require().False(s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice))

	// Test case: Time after thresholdTime and assignedTime
	assignedTime := calculateAssignedTime(
		s.Signaller.valAddress,
		feed.Interval,
		valPrice.Timestamp,
		s.Signaller.distributionStartPercentage,
		s.Signaller.distributionOffsetPercentage,
	)
	s.Signaller.currentBlockTime = assignedTime.Add(time.Second)
	s.Require().True(s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice))

	// Test case: SignalPriceStatus changed, should update
	s.Signaller.currentBlockTime = assignedTime.Add(-time.Second)

	newPrice.Status = feeds.SIGNAL_PRICE_STATUS_AVAILABLE
	s.Require().False(s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice))
	newPrice.Status = feeds.SIGNAL_PRICE_STATUS_UNAVAILABLE
	s.Require().True(s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice))

	// Test case: Price deviated
	s.Signaller.currentBlockTime = assignedTime.Add(-time.Second)
	newPrice.Status = feeds.SIGNAL_PRICE_STATUS_AVAILABLE

	newPrice.Price = 11000 // More than deviationBasisPoint
	s.Require().True(s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice))

	newPrice.Price = 10025 // Within deviationBasisPoint
	s.Require().False(s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice))
}
