package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/globalfee/types"
)

func (s *IntegrationTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name: "set invalid authority",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: s.globalfeeKeeper.GetAuthority(),
				Params: types.Params{
					MinimumGasPrices: sdk.NewDecCoins(
						sdk.NewDecCoin("ALX", sdk.NewInt(1)),
						sdk.NewDecCoinFromDec("BLX", sdk.NewDecWithPrec(1, 3)),
					),
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
