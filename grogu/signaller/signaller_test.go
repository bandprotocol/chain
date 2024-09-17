package signaller

import (
	"sync"
	"testing"
	"time"

	proto "github.com/bandprotocol/bothan/bothan-api/client/go-client/query"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/grogu/signaller/testutil"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	feeds "github.com/bandprotocol/chain/v2/x/feeds/types"
)

type SignallerTestSuite struct {
	suite.Suite

	Signaller    *Signaller
	SubmitCh     chan []feeds.SignalPrice
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
				PriceStatus: feeds.PriceStatusAvailable,
				Validator:   validAddress.String(),
				SignalID:    "signal1",
				Price:       10000000000000,
				Timestamp:   0,
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
			},
		}}, nil).
		AnyTimes()

	mockBothanClient := testutil.NewMockBothanClient(ctrl)
	mockBothanClient.EXPECT().QueryPrices(gomock.Any()).
		Return([]*proto.PriceData{
			{
				SignalId:    "signal1",
				Price:       "10000",
				PriceStatus: proto.PriceStatus_PRICE_STATUS_AVAILABLE,
			},
		}, nil).
		AnyTimes()

	// Create submit channel
	submitCh := make(chan []feeds.SignalPrice, 300)

	// Initialize logger
	allowLevel, _ := log.AllowLevel("info")
	l := logger.New(allowLevel)

	// Initialize pending signal IDs map
	pendingSignalIDs := sync.Map{}

	// Create signaller instance
	s.Signaller = New(
		mockFeedQuerier,
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
		s.Signaller.distributionOffsetPercentage,
		s.Signaller.distributionStartPercentage,
	)
}

func (s *SignallerTestSuite) TestUpdateInternalVariables() {
	success := s.Signaller.updateInternalVariables()
	s.Require().True(success)
	s.Require().NotNil(s.Signaller.params)
	s.Require().NotEmpty(s.Signaller.signalIDToFeed)
	s.Require().NotEmpty(s.Signaller.signalIDToValidatorPrice)
}

func (s *SignallerTestSuite) TestFilterAndPrepareSubmitPrices() {
	s.TestUpdateInternalVariables()

	// Test with available price
	prices := []*proto.PriceData{
		{
			SignalId:    "signal1",
			Price:       "10000",
			PriceStatus: proto.PriceStatus_PRICE_STATUS_AVAILABLE,
		},
	}

	signalIDs := []string{"signal1"}
	// Test with time in the middle of the interval
	middleIntervalTime := time.Unix(30, 0)

	submitPrices := s.Signaller.filterAndPrepareSubmitPrices(prices, signalIDs, middleIntervalTime)
	s.Require().Empty(submitPrices)

	// Test with time at the end of the interval
	endIntervalTime := time.Unix(60, 0)

	submitPrices = s.Signaller.filterAndPrepareSubmitPrices(prices, signalIDs, endIntervalTime)
	s.Require().NotEmpty(submitPrices)
	s.Require().Equal("signal1", submitPrices[0].SignalID)
	s.Require().Equal(uint64(10000*Multiplier), submitPrices[0].Price)

	// Test with unavailable price
	prices = []*proto.PriceData{
		{
			SignalId:    "signal1",
			Price:       "10000",
			PriceStatus: proto.PriceStatus_PRICE_STATUS_UNAVAILABLE,
		},
	}

	// Test with time after the urgent deadline
	afterUrgentDeadlineTime := time.Unix(51, 0)
	submitPrices = s.Signaller.filterAndPrepareSubmitPrices(prices, signalIDs, afterUrgentDeadlineTime)
	s.Require().NotEmpty(submitPrices)
	s.Require().Equal("signal1", submitPrices[0].SignalID)
	s.Require().Equal(uint64(0), submitPrices[0].Price)

	// Test with time before the urgent deadline
	beforeUrgentDeadlineTime := time.Unix(49, 0)
	submitPrices = s.Signaller.filterAndPrepareSubmitPrices(prices, signalIDs, beforeUrgentDeadlineTime)
	s.Require().Empty(submitPrices)
}

func (s *SignallerTestSuite) TestGetAllSignalIDs() {
	signalIDs := s.Signaller.getAllSignalIDs()
	s.Require().Empty(signalIDs)

	// Update internal variables
	s.TestUpdateInternalVariables()

	signalIDs = s.Signaller.getAllSignalIDs()
	s.Require().NotEmpty(signalIDs)
	s.Require().Equal("signal1", signalIDs[0])
}

func (s *SignallerTestSuite) TestGetNonPendingSignalIDs() {
	signalIDs := s.Signaller.getNonPendingSignalIDs()
	s.Require().Empty(signalIDs)

	// Update internal variables
	s.TestUpdateInternalVariables()

	signalIDs = s.Signaller.getNonPendingSignalIDs()
	s.Require().NotEmpty(signalIDs)
	s.Require().Equal("signal1", signalIDs[0])
}

func (s *SignallerTestSuite) TestSubmitPrices() {
	prices := []feeds.SignalPrice{
		{
			SignalID:    "signal1",
			Price:       10000 * Multiplier,
			PriceStatus: feeds.PriceStatusAvailable,
		},
	}

	s.Signaller.submitPrices(prices)

	select {
	case submittedPrices := <-s.SubmitCh:
		s.Require().NotEmpty(submittedPrices)
		s.Require().Equal("signal1", submittedPrices[0].SignalID)
	default:
		s.Fail("Expected prices to be submitted")
	}
}

func (s *SignallerTestSuite) TestIsPriceValid() {
	// Update internal variables
	s.TestUpdateInternalVariables()

	priceData := &proto.PriceData{
		SignalId:    "signal1",
		Price:       "10000",
		PriceStatus: proto.PriceStatus_PRICE_STATUS_AVAILABLE,
	}

	// Test with time before the assigned time
	beforeAssignedTime := time.Unix(s.assignedTime.Unix()-1, 0)
	isValid := s.Signaller.isPriceValid(priceData, beforeAssignedTime)
	s.Require().False(isValid)

	// Test with time at the assigned time
	isValid = s.Signaller.isPriceValid(priceData, s.assignedTime)
	s.Require().True(isValid)

	// Test with time at the start of the interval
	startOfInterval := time.Unix(0, 0)
	isValid = s.Signaller.isPriceValid(priceData, startOfInterval)
	s.Require().False(isValid)
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
		SignalID:  "signal1",
		Price:     10000 * Multiplier,
		Timestamp: 0,
	}

	// Test with new price positive deviation
	thresholdTime := time.Unix(valPrice.Timestamp+s.Signaller.params.CooldownTime+TimeBuffer, 0)
	newPrice := uint64(10050 * Multiplier)

	shouldUpdate := s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice, thresholdTime)
	s.Require().True(shouldUpdate)

	// Test with new price negative deviation
	newPrice = uint64(9950 * Multiplier)

	shouldUpdate = s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice, thresholdTime)
	s.Require().True(shouldUpdate)

	// Test with new price within deviation
	newPrice = uint64(10025 * Multiplier)

	shouldUpdate = s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice, thresholdTime)
	s.Require().False(shouldUpdate)

	// Test with new price outside deviation
	newPrice = uint64(10075 * Multiplier)

	shouldUpdate = s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice, thresholdTime)
	s.Require().True(shouldUpdate)

	// Test with time before threshold time, price outside deviation
	newPrice = uint64(10075 * Multiplier)
	beforeThresholdTime := time.Unix(valPrice.Timestamp+s.Signaller.params.CooldownTime, 0)

	shouldUpdate = s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice, beforeThresholdTime)
	s.Require().False(shouldUpdate)

	// Test with time at assigned time, price within deviation
	newPrice = uint64(10025 * Multiplier)

	shouldUpdate = s.Signaller.shouldUpdatePrice(feed, valPrice, newPrice, s.assignedTime)
	s.Require().True(shouldUpdate)
}