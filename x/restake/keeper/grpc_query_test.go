package keeper_test

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestQueryKeys() {
	ctx, queryClient := suite.ctx, suite.queryClient

	var validKeys []*types.Key
	for i, key := range suite.validKeys {
		suite.restakeKeeper.SetKey(ctx, key)
		validKeys = append(validKeys, &suite.validKeys[i])
	}

	// query and check
	var (
		req    *types.QueryKeysRequest
		expRes *types.QueryKeysResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"all keys",
			func() {
				req = &types.QueryKeysRequest{}
				expRes = &types.QueryKeysResponse{
					Keys: validKeys,
				}
			},
			true,
		},
		{
			"limit 1",
			func() {
				req = &types.QueryKeysRequest{
					Pagination: &query.PageRequest{Limit: 1},
				}
				expRes = &types.QueryKeysResponse{
					Keys: validKeys[:1],
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Keys(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetKeys(), res.GetKeys())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryKey() {
	queryClient := suite.queryClient
	suite.setupState()

	// query and check
	res, err := queryClient.Key(context.Background(), &types.QueryKeyRequest{
		Key: KeyWithRewards,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryKeyResponse{
		Key: suite.validKeys[0],
	}, res)

	res, err = queryClient.Key(context.Background(), &types.QueryKeyRequest{
		Key: InvalidKey,
	})
	suite.Require().ErrorContains(err, "key not found")
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestQueryRewards() {
	queryClient := suite.queryClient
	suite.setupState()

	// query and check
	var (
		req    *types.QueryRewardsRequest
		expRes *types.QueryRewardsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"rewards of address1 - lock on multiple keys",
			func() {
				req = &types.QueryRewardsRequest{
					LockerAddress: ValidAddress1.String(),
				}
				expRes = &types.QueryRewardsResponse{
					Rewards: []*types.Reward{
						{
							Key:     KeyWithRewards,
							Rewards: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
						},
						{
							Key:     KeyWithoutRewards,
							Rewards: nil,
						},
						{
							Key:     InactiveKey,
							Rewards: nil,
						},
					},
				}
			},
			true,
		},
		{
			"rewards of address2 - lock on one key",
			func() {
				req = &types.QueryRewardsRequest{
					LockerAddress: ValidAddress2.String(),
				}
				expRes = &types.QueryRewardsResponse{
					Rewards: []*types.Reward{
						{
							Key:     KeyWithRewards,
							Rewards: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
						},
					},
				}
			},
			true,
		},
		{
			"rewards of address3 - no lock",
			func() {
				req = &types.QueryRewardsRequest{
					LockerAddress: ValidAddress3.String(),
				}
				expRes = &types.QueryRewardsResponse{
					Rewards: []*types.Reward(nil),
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Rewards(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetRewards(), res.GetRewards())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryLocks() {
	queryClient := suite.queryClient
	suite.setupState()

	// query and check
	var (
		req    *types.QueryLocksRequest
		expRes *types.QueryLocksResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"locks of address1 - lock on multiple keys",
			func() {
				req = &types.QueryLocksRequest{
					LockerAddress: ValidAddress1.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse{
						{
							Key:    KeyWithRewards,
							Amount: sdkmath.NewInt(10),
						},
						{
							Key:    KeyWithoutRewards,
							Amount: sdkmath.NewInt(100),
						},
					},
				}
			},
			true,
		},
		{
			"locks of address2 - lock on one key",
			func() {
				req = &types.QueryLocksRequest{
					LockerAddress: ValidAddress2.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse{
						{
							Key:    KeyWithRewards,
							Amount: sdkmath.NewInt(10),
						},
					},
				}
			},
			true,
		},
		{
			"locks of address3 - no lock",
			func() {
				req = &types.QueryLocksRequest{
					LockerAddress: ValidAddress3.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse(nil),
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Locks(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetLocks(), res.GetLocks())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}
