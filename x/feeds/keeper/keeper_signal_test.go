package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetVote() {
	ctx := suite.ctx

	// set
	expVote := types.Vote{
		Voter: ValidVoter.String(),
		Signals: []types.Signal{
			{
				ID:    "CS:BAND-USD",
				Power: 1e9,
			},
			{
				ID:    "CS:BTC-USD",
				Power: 1e9,
			},
		},
	}
	suite.feedsKeeper.SetVote(ctx, expVote)

	// get
	signals := suite.feedsKeeper.GetVote(ctx, ValidVoter)
	suite.Require().Equal(expVote.Signals, signals)
}

func (suite *KeeperTestSuite) TestGetSetDeleteSignalTotalPower() {
	ctx := suite.ctx

	// set
	expSignalTotalPower := types.Signal{
		ID:    "CS:BAND-USD",
		Power: 1e9,
	}
	suite.feedsKeeper.SetSignalTotalPower(ctx, expSignalTotalPower)

	// get
	signal, err := suite.feedsKeeper.GetSignalTotalPower(ctx, expSignalTotalPower.ID)
	suite.Require().NoError(err)
	suite.Require().Equal(expSignalTotalPower, signal)

	// set with power 0
	SignalTotalPowerZero := types.Signal{
		ID:    "CS:BAND-USD",
		Power: 0,
	}
	suite.feedsKeeper.SetSignalTotalPower(ctx, SignalTotalPowerZero)

	// get
	signal, err = suite.feedsKeeper.GetSignalTotalPower(ctx, SignalTotalPowerZero.ID)
	suite.Require().Error(err)
	suite.Require().Equal(types.Signal{}, signal)
}

func (suite *KeeperTestSuite) TestUpdateVoteAndReturnPowerDiff() {
	ctx := suite.ctx

	tests := []struct {
		name              string
		previousVote      *types.Vote
		voter             sdk.AccAddress
		signals           []types.Signal
		expectedPowerDiff map[string]int64
	}{
		{
			name:         "no previous vote",
			previousVote: nil,
			voter:        ValidVoter,
			signals: []types.Signal{
				{
					ID:    "CS:BAND-USD",
					Power: 1000,
				},
			},
			expectedPowerDiff: map[string]int64{
				"CS:BAND-USD": 1000,
			},
		},
		{
			name: "empty previous vote",
			previousVote: &types.Vote{
				Voter:   ValidVoter.String(),
				Signals: []types.Signal{},
			},
			voter: ValidVoter,
			signals: []types.Signal{
				{
					ID:    "CS:BAND-USD",
					Power: 1000,
				},
			},
			expectedPowerDiff: map[string]int64{
				"CS:BAND-USD": 1000,
			},
		},
		{
			name: "vote more than previous",
			previousVote: &types.Vote{
				Voter: ValidVoter.String(),
				Signals: []types.Signal{
					{
						ID:    "CS:BAND-USD",
						Power: 1000,
					},
				},
			},
			voter: ValidVoter,
			signals: []types.Signal{
				{
					ID:    "CS:BAND-USD",
					Power: 3000,
				},
			},
			expectedPowerDiff: map[string]int64{
				"CS:BAND-USD": 2000,
			},
		},
		{
			name: "vote less than previous",
			previousVote: &types.Vote{
				Voter: ValidVoter.String(),
				Signals: []types.Signal{
					{
						ID:    "CS:BAND-USD",
						Power: 1000,
					},
				},
			},
			voter: ValidVoter,
			signals: []types.Signal{
				{
					ID:    "CS:BAND-USD",
					Power: 500,
				},
			},
			expectedPowerDiff: map[string]int64{
				"CS:BAND-USD": -500,
			},
		},
		{
			name: "empty new vote",
			previousVote: &types.Vote{
				Voter: ValidVoter.String(),
				Signals: []types.Signal{
					{
						ID:    "CS:BAND-USD",
						Power: 1000,
					},
				},
			},
			voter:   ValidVoter,
			signals: []types.Signal{},
			expectedPowerDiff: map[string]int64{
				"CS:BAND-USD": -1000,
			},
		},
		{
			name: "multiple signals",
			previousVote: &types.Vote{
				Voter: ValidVoter.String(),
				Signals: []types.Signal{
					{
						ID:    "CS:BAND-USD",
						Power: 1000,
					},
					{
						ID:    "CS:ETH-USD",
						Power: 2000,
					},
					{
						ID:    "CS:BTC-USD",
						Power: 3000,
					},
				},
			},
			voter: ValidVoter,
			signals: []types.Signal{
				{
					ID:    "CS:BAND-USD",
					Power: 2000,
				},
				{
					ID:    "CS:ETH-USD",
					Power: 1000,
				},
				{
					ID:    "CS:ATOM-USD",
					Power: 600,
				},
			},
			expectedPowerDiff: map[string]int64{
				"CS:BAND-USD": 1000,
				"CS:ETH-USD":  -1000,
				"CS:BTC-USD":  -3000,
				"CS:ATOM-USD": 600,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.previousVote != nil {
				suite.feedsKeeper.SetVote(ctx, *tt.previousVote)
			}

			powerDiff := suite.feedsKeeper.UpdateVoteAndReturnPowerDiff(
				ctx,
				tt.voter,
				tt.signals,
			)

			suite.Require().Equal(len(tt.expectedPowerDiff), len(powerDiff))
			for key, val := range powerDiff {
				suite.Require().Equal(tt.expectedPowerDiff[key], val)
			}
		})
	}
}
