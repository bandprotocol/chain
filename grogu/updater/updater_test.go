package updater

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	"github.com/bandprotocol/chain/v3/grogu/updater/testutil"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	feeds "github.com/bandprotocol/chain/v3/x/feeds/types"
)

type UpdaterTestSuite struct {
	suite.Suite

	Updater *Updater
}

func TestUpdaterTestSuite(t *testing.T) {
	suite.Run(t, new(UpdaterTestSuite))
}

func (s *UpdaterTestSuite) SetupTest() {
	// Set up mock types
	ctrl := gomock.NewController(s.T())
	mockFeedQuerier := testutil.NewMockFeedQuerier(ctrl)
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
	mockFeedQuerier.EXPECT().QueryReferenceSourceConfig().Return(&feeds.QueryReferenceSourceConfigResponse{
		ReferenceSourceConfig: feeds.DefaultReferenceSourceConfig(),
	}, nil).AnyTimes()

	mockBothanClient := testutil.NewMockBothanClient(ctrl)
	mockBothanClient.EXPECT().SetActiveSignalIDs(gomock.Any()).
		Return(nil).
		AnyTimes()
	mockBothanClient.EXPECT().UpdateRegistry(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	mockClient := testutil.NewMockRemoteClient(ctrl)
	mockClient.EXPECT().Remote().Return("mock").AnyTimes()
	mockClient.EXPECT().Subscribe(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(make(chan coretypes.ResultEvent), nil).
		AnyTimes()
	mockRPCClients := []rpcclient.RemoteClient{mockClient}

	// Initialize logger
	allowLevel, _ := log.ParseLogLevel("info")
	l := logger.NewLogger(allowLevel)

	// Initialize max heights
	maxCurrentFeedEventHeight := new(atomic.Int64)
	maxCurrentFeedEventHeight.Store(0)

	maxUpdateRefSourceEventHeight := new(atomic.Int64)
	maxUpdateRefSourceEventHeight.Store(0)

	// Set up updater
	s.Updater = New(
		mockFeedQuerier,
		mockBothanClient,
		mockRPCClients,
		l,
		maxCurrentFeedEventHeight,
		maxUpdateRefSourceEventHeight,
	)
}

func (s *UpdaterTestSuite) TestUpdaterInit() {
	err := s.Updater.Init()
	s.Require().NoError(err)
}

func (s *UpdaterTestSuite) TestUpdateBothanActiveFeeds() {
	err := s.Updater.updateBothanActiveFeeds()
	s.Require().NoError(err)
}

func (s *UpdaterTestSuite) TestUpdateBothanRegistry() {
	err := s.Updater.updateBothanRegistry()
	s.Require().NoError(err)
}

func (s *UpdaterTestSuite) TestProcessEventCurrentFeeds() {
	event := coretypes.ResultEvent{
		Data: tmtypes.EventDataNewBlock{
			Block: &tmtypes.Block{Header: tmtypes.Header{Height: 10}},
		},
	}

	processEvent(
		event,
		s.Updater.logger,
		CurrentFeedsQuery,
		s.Updater.maxCurrentFeedsEventHeight,
		func(ev coretypes.ResultEvent) int64 {
			return ev.Data.(tmtypes.EventDataNewBlock).Block.Height
		},
		s.Updater.updateBothanActiveFeeds,
	)

	s.Require().Equal(int64(10), s.Updater.maxCurrentFeedsEventHeight.Load())
}

func (s *UpdaterTestSuite) TestProcessEventUpdateRefSource() {
	event := coretypes.ResultEvent{
		Data: tmtypes.EventDataTx{
			TxResult: abci.TxResult{
				Height: 10,
			},
		},
	}

	processEvent(
		event,
		s.Updater.logger,
		UpdateRefSourceQuery,
		s.Updater.maxUpdateRefSourceEventHeight,
		func(ev coretypes.ResultEvent) int64 {
			return ev.Data.(tmtypes.EventDataTx).TxResult.Height
		},
		s.Updater.updateBothanRegistry,
	)

	s.Require().Equal(int64(10), s.Updater.maxUpdateRefSourceEventHeight.Load())
}
