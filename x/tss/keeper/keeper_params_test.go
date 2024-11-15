package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetParams() {
	ctx, k := s.ctx, s.keeper
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	s.Require().NoError(err)

	s.Require().Equal(params, k.GetParams(ctx))
}

func (s *KeeperTestSuite) TestParams() {
	k := s.keeper

	testCases := []struct {
		name         string
		input        types.Params
		expectErr    bool
		expectErrStr string
	}{
		{
			name: "set invalid params",
			input: types.Params{
				MaxGroupSize:      0,
				MaxDESize:         0,
				CreationPeriod:    1,
				SigningPeriod:     1,
				MaxSigningAttempt: 1,
				MaxMemoLength:     1,
				MaxMessageLength:  1,
			},
			expectErr:    true,
			expectErrStr: "must be positive:",
		},
		{
			name: "set full valid params",
			input: types.Params{
				MaxGroupSize:      types.DefaultMaxGroupSize,
				MaxDESize:         types.DefaultMaxDESize,
				CreationPeriod:    types.DefaultCreationPeriod,
				SigningPeriod:     types.DefaultSigningPeriod,
				MaxSigningAttempt: types.DefaultMaxSigningAttempt,
				MaxMemoLength:     types.DefaultMaxMemoLength,
				MaxMessageLength:  types.DefaultMaxMessageLength,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expected := k.GetParams(s.ctx)
			err := k.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErrStr)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := k.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}
