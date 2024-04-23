package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestAfterSigningFailed() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	hook := k.Hooks()

	testCases := []struct {
		name           string
		signing        tsstypes.Signing
		bandtssSignign types.Signing
		expCoins       sdk.Coins
	}{
		{
			"10uband with 2 members",
			tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			types.Signing{
				ID:                    1,
				Fee:                   sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				CurrentGroupSigningID: 1,
				Requester:             testapp.FeePayer.Address.String(),
			},
			sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
		},
		{
			"10uband,15token with 2 members",
			tsstypes.Signing{
				ID:      2,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			types.Signing{
				ID:                    2,
				Fee:                   sdk.NewCoins(sdk.NewInt64Coin("uband", 10), sdk.NewInt64Coin("token", 15)),
				CurrentGroupSigningID: 2,
				Requester:             testapp.FeePayer.Address.String(),
			},
			sdk.NewCoins(sdk.NewInt64Coin("uband", 20), sdk.NewInt64Coin("token", 30)),
		},
		{
			"0uband with 2 members",
			tsstypes.Signing{
				ID:      3,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			types.Signing{
				ID:                    3,
				Fee:                   sdk.NewCoins(),
				CurrentGroupSigningID: 3,
				Requester:             testapp.FeePayer.Address.String(),
			},
			sdk.NewCoins(),
		},
		{
			"10uband with 0 member",
			tsstypes.Signing{
				ID:              4,
				GroupID:         1,
				AssignedMembers: []tsstypes.AssignedMember{},
			},
			types.Signing{
				ID:                    4,
				Fee:                   sdk.NewCoins(),
				CurrentGroupSigningID: 4,
				Requester:             testapp.FeePayer.Address.String(),
			},
			sdk.NewCoins(),
		},
	}

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			k.SetSigning(ctx, tc.bandtssSignign)
			k.SetSigningIDMapping(ctx, tc.signing.ID, tc.bandtssSignign.ID)

			balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
			balancesModuleBefore := s.app.BankKeeper.GetAllBalances(ctx, k.GetBandtssAccount(ctx).GetAddress())

			err := hook.AfterSigningFailed(ctx, tc.signing)
			s.Require().NoError(err)

			balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
			balancesModuleAfter := s.app.BankKeeper.GetAllBalances(ctx, k.GetBandtssAccount(ctx).GetAddress())

			gain := balancesAfter.Sub(balancesBefore...)
			s.Require().Equal(tc.expCoins, gain)

			lose := balancesModuleBefore.Sub(balancesModuleAfter...)
			s.Require().Equal(tc.expCoins, lose)
		})
	}
}
