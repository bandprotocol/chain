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
					StakerAddress: ValidAddress1.String(),
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
					StakerAddress: ValidAddress2.String(),
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
					StakerAddress: ValidAddress3.String(),
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

func (suite *KeeperTestSuite) TestQueryReward() {
	queryClient := suite.queryClient
	suite.setupState()

	// query and check
	var (
		req    *types.QueryRewardRequest
		expRes *types.QueryRewardResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"reward of address1 on KeyWithRewards",
			func() {
				req = &types.QueryRewardRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           KeyWithRewards,
				}
				expRes = &types.QueryRewardResponse{
					Reward: types.Reward{
						Key:     KeyWithRewards,
						Rewards: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
					},
				}
			},
			true,
		},
		{
			"reward of address1 on InactiveKey",
			func() {
				req = &types.QueryRewardRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           InactiveKey,
				}
				expRes = &types.QueryRewardResponse{
					Reward: types.Reward{
						Key:     InactiveKey,
						Rewards: nil,
					},
				}
			},
			true,
		},
		{
			"reward of address2 on KeyWithRewards",
			func() {
				req = &types.QueryRewardRequest{
					StakerAddress: ValidAddress2.String(),
					Key:           KeyWithRewards,
				}
				expRes = &types.QueryRewardResponse{
					Reward: types.Reward{
						Key:     KeyWithRewards,
						Rewards: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
					},
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Reward(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetReward(), res.GetReward())
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
					StakerAddress: ValidAddress1.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse{
						{
							Key:   KeyWithRewards,
							Power: sdkmath.NewInt(10),
						},
						{
							Key:   KeyWithoutRewards,
							Power: sdkmath.NewInt(100),
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
					StakerAddress: ValidAddress2.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse{
						{
							Key:   KeyWithRewards,
							Power: sdkmath.NewInt(10),
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
					StakerAddress: ValidAddress3.String(),
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

func (suite *KeeperTestSuite) TestQueryLock() {
	queryClient := suite.queryClient
	suite.setupState()

	// query and check
	var (
		req    *types.QueryLockRequest
		expRes *types.QueryLockResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"lock of address1 on KeyWithRewards",
			func() {
				req = &types.QueryLockRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           KeyWithRewards,
				}
				expRes = &types.QueryLockResponse{
					Lock: types.LockResponse{
						Key:   KeyWithRewards,
						Power: sdk.NewInt(10),
					},
				}
			},
			true,
		},
		{
			"lock of address1 on InactiveKey",
			func() {
				req = &types.QueryLockRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           InactiveKey,
				}
				expRes = nil
			},
			false,
		},
		{
			"lock of address2 on KeyWithRewards",
			func() {
				req = &types.QueryLockRequest{
					StakerAddress: ValidAddress2.String(),
					Key:           KeyWithRewards,
				}
				expRes = &types.QueryLockResponse{
					Lock: types.LockResponse{
						Key:   KeyWithRewards,
						Power: sdk.NewInt(10),
					},
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Lock(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetLock(), res.GetLock())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}
