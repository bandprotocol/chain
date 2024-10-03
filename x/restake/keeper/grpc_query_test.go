package keeper_test

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

func (suite *KeeperTestSuite) TestQueryVaults() {
	ctx, queryClient := suite.ctx, suite.queryClient

	var validVaults []*types.Vault
	for i, vault := range suite.validVaults {
		suite.restakeKeeper.SetVault(ctx, vault)
		validVaults = append(validVaults, &suite.validVaults[i])
	}

	// query and check
	var (
		req    *types.QueryVaultsRequest
		expRes *types.QueryVaultsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"all vaults",
			func() {
				req = &types.QueryVaultsRequest{}
				expRes = &types.QueryVaultsResponse{
					Vaults: validVaults,
				}
			},
			true,
		},
		{
			"limit 1",
			func() {
				req = &types.QueryVaultsRequest{
					Pagination: &query.PageRequest{Limit: 1},
				}
				expRes = &types.QueryVaultsResponse{
					Vaults: validVaults[:1],
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Vaults(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetVaults(), res.GetVaults())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryVault() {
	queryClient := suite.queryClient
	suite.setupState()

	// query and check
	res, err := queryClient.Vault(context.Background(), &types.QueryVaultRequest{
		Key: VaultKeyWithRewards,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryVaultResponse{
		Vault: suite.validVaults[0],
	}, res)

	res, err = queryClient.Vault(context.Background(), &types.QueryVaultRequest{
		Key: InvalidVaultKey,
	})
	suite.Require().ErrorContains(err, "vault not found")
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
			"rewards of address1 - lock on multiple vaults",
			func() {
				req = &types.QueryRewardsRequest{
					StakerAddress: ValidAddress1.String(),
				}
				expRes = &types.QueryRewardsResponse{
					Rewards: []*types.Reward{
						{
							Key:     VaultKeyWithRewards,
							Rewards: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
						},
						{
							Key:     VaultKeyWithoutRewards,
							Rewards: nil,
						},
						{
							Key:     InactiveVaultKey,
							Rewards: nil,
						},
					},
				}
			},
			true,
		},
		{
			"rewards of address2 - lock on one vault",
			func() {
				req = &types.QueryRewardsRequest{
					StakerAddress: ValidAddress2.String(),
				}
				expRes = &types.QueryRewardsResponse{
					Rewards: []*types.Reward{
						{
							Key:     VaultKeyWithRewards,
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
			"reward of address1 on VaultKeyWithRewards",
			func() {
				req = &types.QueryRewardRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           VaultKeyWithRewards,
				}
				expRes = &types.QueryRewardResponse{
					Reward: types.Reward{
						Key:     VaultKeyWithRewards,
						Rewards: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
					},
				}
			},
			true,
		},
		{
			"reward of address1 on InactiveVaultKey",
			func() {
				req = &types.QueryRewardRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           InactiveVaultKey,
				}
				expRes = &types.QueryRewardResponse{
					Reward: types.Reward{
						Key:     InactiveVaultKey,
						Rewards: nil,
					},
				}
			},
			true,
		},
		{
			"reward of address2 on VaultKeyWithRewards",
			func() {
				req = &types.QueryRewardRequest{
					StakerAddress: ValidAddress2.String(),
					Key:           VaultKeyWithRewards,
				}
				expRes = &types.QueryRewardResponse{
					Reward: types.Reward{
						Key:     VaultKeyWithRewards,
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
			"locks of address1 - lock on multiple vaults",
			func() {
				req = &types.QueryLocksRequest{
					StakerAddress: ValidAddress1.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse{
						{
							Key:   VaultKeyWithRewards,
							Power: sdkmath.NewInt(10),
						},
						{
							Key:   VaultKeyWithoutRewards,
							Power: sdkmath.NewInt(100),
						},
					},
				}
			},
			true,
		},
		{
			"locks of address2 - lock on one vault",
			func() {
				req = &types.QueryLocksRequest{
					StakerAddress: ValidAddress2.String(),
				}
				expRes = &types.QueryLocksResponse{
					Locks: []*types.LockResponse{
						{
							Key:   VaultKeyWithRewards,
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
			"lock of address1 on VaultKeyWithRewards",
			func() {
				req = &types.QueryLockRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           VaultKeyWithRewards,
				}
				expRes = &types.QueryLockResponse{
					Lock: types.LockResponse{
						Key:   VaultKeyWithRewards,
						Power: sdkmath.NewInt(10),
					},
				}
			},
			true,
		},
		{
			"lock of address1 on InactiveVaultKey",
			func() {
				req = &types.QueryLockRequest{
					StakerAddress: ValidAddress1.String(),
					Key:           InactiveVaultKey,
				}
				expRes = nil
			},
			false,
		},
		{
			"lock of address2 on VaultKeyWithRewards",
			func() {
				req = &types.QueryLockRequest{
					StakerAddress: ValidAddress2.String(),
					Key:           VaultKeyWithRewards,
				}
				expRes = &types.QueryLockResponse{
					Lock: types.LockResponse{
						Key:   VaultKeyWithRewards,
						Power: sdkmath.NewInt(10),
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
