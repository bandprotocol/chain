package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestGetOrCreateKey() {
	ctx := suite.ctx
	suite.setupState()

	// already existed keys
	for _, expKey := range suite.validKeys {
		key, err := suite.restakeKeeper.GetOrCreateKey(ctx, expKey.Name)
		suite.Require().NoError(err)
		suite.Require().Equal(expKey, key)
	}

	// new key
	key, err := suite.restakeKeeper.GetOrCreateKey(ctx, "newKey")
	suite.Require().NoError(err)
	suite.Require().Equal(types.Key{
		Name:            "newKey",
		PoolAddress:     "cosmos1x9lj2q3l80xfljcuuw89grm6jw96txayk9z8m0q4g658xe789dxszl8a6s",
		IsActive:        true,
		RewardPerPowers: sdk.NewDecCoins(),
		TotalPower:      sdkmath.NewInt(0),
		Remainders:      sdk.NewDecCoins(),
	}, key)
}

func (suite *KeeperTestSuite) TestAddRewards() {
	ctx := suite.ctx
	suite.setupState()

	// error case -  no key
	err := suite.restakeKeeper.AddRewards(
		ctx,
		RewarderAddress,
		InvalidKey,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
	)
	suite.Require().Error(err)

	// error case - key is deactivated
	err = suite.restakeKeeper.AddRewards(
		ctx,
		RewarderAddress,
		InactiveKey,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
	)
	suite.Require().Error(err)

	// error case - total power is zero
	err = suite.restakeKeeper.AddRewards(
		ctx,
		RewarderAddress,
		KeyWithoutLocks,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
	)
	suite.Require().Error(err)

	// success cases
	var (
		key     types.Key
		rewards sdk.Coins
	)

	testCases := []struct {
		name     string
		malleate func()
		expKey   types.Key
	}{
		{
			"success case - 1 coin",
			func() {
				key = types.Key{
					Name:            KeyWithRewards,
					PoolAddress:     KeyWithRewardsPoolAddress.String(),
					IsActive:        true,
					RewardPerPowers: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
				)
			},
			types.Key{
				Name:        KeyWithRewards,
				PoolAddress: KeyWithRewardsPoolAddress.String(),
				IsActive:    true,
				RewardPerPowers: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 2)),
				),
				TotalPower: sdkmath.NewInt(100),
				Remainders: nil,
			},
		},
		{
			"success case - 2 coin same amount",
			func() {
				key = types.Key{
					Name:            KeyWithRewards,
					PoolAddress:     KeyWithRewardsPoolAddress.String(),
					IsActive:        true,
					RewardPerPowers: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1)),
				)
			},
			types.Key{
				Name:        KeyWithRewards,
				PoolAddress: KeyWithRewardsPoolAddress.String(),
				IsActive:    true,
				RewardPerPowers: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 2)),
					sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
				),
				TotalPower: sdkmath.NewInt(100),
				Remainders: nil,
			},
		},
		{
			"success case - 2 coin diff amount",
			func() {
				key = types.Key{
					Name:            KeyWithRewards,
					PoolAddress:     KeyWithRewardsPoolAddress.String(),
					IsActive:        true,
					RewardPerPowers: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(5)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1)),
				)
			},
			types.Key{
				Name:        KeyWithRewards,
				PoolAddress: KeyWithRewardsPoolAddress.String(),
				IsActive:    true,
				RewardPerPowers: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(5, 2)),
					sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
				),
				TotalPower: sdkmath.NewInt(100),
				Remainders: nil,
			},
		},
		{
			"success case - small reward, big total power",
			func() {
				key = types.Key{
					Name:            KeyWithRewards,
					PoolAddress:     KeyWithRewardsPoolAddress.String(),
					IsActive:        true,
					RewardPerPowers: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(1e18),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
				)
			},
			types.Key{
				Name:        KeyWithRewards,
				PoolAddress: KeyWithRewardsPoolAddress.String(),
				IsActive:    true,
				RewardPerPowers: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
				),
				TotalPower: sdkmath.NewInt(1e18),
				Remainders: nil,
			},
		},
		{
			"success case - big reward, small total power",
			func() {
				key = types.Key{
					Name:            KeyWithRewards,
					PoolAddress:     KeyWithRewardsPoolAddress.String(),
					IsActive:        true,
					RewardPerPowers: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(1),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1e18)),
				)
			},
			types.Key{
				Name:        KeyWithRewards,
				PoolAddress: KeyWithRewardsPoolAddress.String(),
				IsActive:    true,
				RewardPerPowers: sdk.NewDecCoins(
					sdk.NewDecCoin("aaaa", sdkmath.NewInt(1e18)),
				),
				TotalPower: sdkmath.NewInt(1),
				Remainders: nil,
			},
		},
		{
			"success case - have remainder",
			func() {
				key = types.Key{
					Name:            KeyWithRewards,
					PoolAddress:     KeyWithRewardsPoolAddress.String(),
					IsActive:        true,
					RewardPerPowers: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(3),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
				)
			},
			types.Key{
				Name:        KeyWithRewards,
				PoolAddress: KeyWithRewardsPoolAddress.String(),
				IsActive:    true,
				RewardPerPowers: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdk.MustNewDecFromStr("0.333333333333333333")),
				),
				TotalPower: sdkmath.NewInt(3),
				Remainders: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdk.NewDecWithPrec(1, 18)),
				),
			},
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.name), func() {
			suite.resetState()
			ctx = suite.ctx
			testCase.malleate()

			suite.restakeKeeper.SetKey(ctx, key)
			err = suite.restakeKeeper.AddRewards(
				ctx,
				RewarderAddress,
				key.Name,
				rewards,
			)
			suite.Require().NoError(err)

			key, err := suite.restakeKeeper.GetKey(ctx, key.Name)
			suite.Require().NoError(err)
			suite.Require().Equal(testCase.expKey, key)
		})
	}
}

func (suite *KeeperTestSuite) TestIsActiveKey() {
	ctx := suite.ctx
	suite.setupState()

	// case - valid key
	for _, expKey := range suite.validKeys {
		isActive := suite.restakeKeeper.IsActiveKey(ctx, expKey.Name)
		suite.Require().Equal(expKey.IsActive, isActive)
	}

	// case - no key
	isActive := suite.restakeKeeper.IsActiveKey(ctx, InvalidKey)
	suite.Require().Equal(false, isActive)
}

func (suite *KeeperTestSuite) TestDeactivateKey() {
	ctx := suite.ctx
	suite.setupState()

	// error case -  no key
	err := suite.restakeKeeper.DeactivateKey(ctx, InvalidKey)
	suite.Require().Error(err)

	// error case - key is deactivated
	err = suite.restakeKeeper.DeactivateKey(ctx, InactiveKey)
	suite.Require().Error(err)

	// success case
	err = suite.restakeKeeper.DeactivateKey(ctx, KeyWithRewards)
	suite.Require().NoError(err)
	key, err := suite.restakeKeeper.GetKey(ctx, KeyWithRewards)
	suite.Require().NoError(err)
	suite.Require().Equal(false, key.IsActive)
}

func (suite *KeeperTestSuite) TestGetSetKey() {
	ctx := suite.ctx

	// set
	expectedKeys := suite.validKeys
	for _, expKey := range expectedKeys {
		suite.restakeKeeper.SetKey(ctx, expKey)

		// has
		has := suite.restakeKeeper.HasKey(ctx, expKey.Name)
		suite.Require().True(has)

		// get
		key, err := suite.restakeKeeper.GetKey(ctx, expKey.Name)
		suite.Require().NoError(err)
		suite.Require().Equal(expKey, key)

		// must get
		key = suite.restakeKeeper.MustGetKey(ctx, expKey.Name)
		suite.Require().Equal(expKey, key)
	}

	// has
	has := suite.restakeKeeper.HasKey(ctx, "nonKey")
	suite.Require().False(has)

	// get
	keys := suite.restakeKeeper.GetKeys(ctx)
	suite.Require().Equal(expectedKeys, keys)

	_, err := suite.restakeKeeper.GetKey(ctx, "nonKey")
	suite.Require().Error(err)

	// must get
	suite.Require().Panics(func() {
		_ = suite.restakeKeeper.MustGetKey(ctx, "nonKey")
	})
}
