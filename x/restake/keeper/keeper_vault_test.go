package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestGetOrCreateVault() {
	ctx := suite.ctx
	suite.setupState()

	// already existed vaults
	for _, expVault := range suite.validVaults {
		vault, err := suite.restakeKeeper.GetOrCreateVault(ctx, expVault.Key)
		suite.Require().NoError(err)
		suite.Require().Equal(expVault, vault)
	}

	// new vault
	vault, err := suite.restakeKeeper.GetOrCreateVault(ctx, "newVault")
	suite.Require().NoError(err)
	suite.Require().Equal(types.Vault{
		Key:             "newVault",
		VaultAddress:    "cosmos19p78qeezm3l7pycx3mjrs5dq0p5znwddndsrkvgt7ewq3qg7vf6q3rr6gl",
		IsActive:        true,
		RewardsPerPower: sdk.NewDecCoins(),
		TotalPower:      sdkmath.NewInt(0),
		Remainders:      sdk.NewDecCoins(),
	}, vault)
}

func (suite *KeeperTestSuite) TestAddRewards() {
	ctx := suite.ctx
	suite.setupState()

	// error case -  no vault
	err := suite.restakeKeeper.AddRewards(
		ctx,
		RewarderAddress,
		InvalidVaultKey,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
	)
	suite.Require().Error(err)

	// error case - vault is deactivated
	err = suite.restakeKeeper.AddRewards(
		ctx,
		RewarderAddress,
		InactiveVaultKey,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
	)
	suite.Require().Error(err)

	// error case - total power is zero
	err = suite.restakeKeeper.AddRewards(
		ctx,
		RewarderAddress,
		VaultKeyWithoutLocks,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
	)
	suite.Require().Error(err)

	// success cases
	var (
		vault   types.Vault
		rewards sdk.Coins
	)

	testCases := []struct {
		name     string
		malleate func()
		expVault types.Vault
	}{
		{
			"success case - 1 coin",
			func() {
				vault = types.Vault{
					Key:             VaultKeyWithRewards,
					VaultAddress:    VaultWithRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
				)
			},
			types.Vault{
				Key:          VaultKeyWithRewards,
				VaultAddress: VaultWithRewardsAddress.String(),
				IsActive:     true,
				RewardsPerPower: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 2)),
				),
				TotalPower: sdkmath.NewInt(100),
				Remainders: nil,
			},
		},
		{
			"success case - 2 coin same amount",
			func() {
				vault = types.Vault{
					Key:             VaultKeyWithRewards,
					VaultAddress:    VaultWithRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1)),
				)
			},
			types.Vault{
				Key:          VaultKeyWithRewards,
				VaultAddress: VaultWithRewardsAddress.String(),
				IsActive:     true,
				RewardsPerPower: sdk.NewDecCoins(
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
				vault = types.Vault{
					Key:             VaultKeyWithRewards,
					VaultAddress:    VaultWithRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(5)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1)),
				)
			},
			types.Vault{
				Key:          VaultKeyWithRewards,
				VaultAddress: VaultWithRewardsAddress.String(),
				IsActive:     true,
				RewardsPerPower: sdk.NewDecCoins(
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
				vault = types.Vault{
					Key:             VaultKeyWithRewards,
					VaultAddress:    VaultWithRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(1e18),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
				)
			},
			types.Vault{
				Key:          VaultKeyWithRewards,
				VaultAddress: VaultWithRewardsAddress.String(),
				IsActive:     true,
				RewardsPerPower: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
				),
				TotalPower: sdkmath.NewInt(1e18),
				Remainders: nil,
			},
		},
		{
			"success case - big reward, small total power",
			func() {
				vault = types.Vault{
					Key:             VaultKeyWithRewards,
					VaultAddress:    VaultWithRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(1),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1e18)),
				)
			},
			types.Vault{
				Key:          VaultKeyWithRewards,
				VaultAddress: VaultWithRewardsAddress.String(),
				IsActive:     true,
				RewardsPerPower: sdk.NewDecCoins(
					sdk.NewDecCoin("aaaa", sdkmath.NewInt(1e18)),
				),
				TotalPower: sdkmath.NewInt(1),
				Remainders: nil,
			},
		},
		{
			"success case - have remainder",
			func() {
				vault = types.Vault{
					Key:             VaultKeyWithRewards,
					VaultAddress:    VaultWithRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: sdk.NewDecCoins(),
					TotalPower:      sdkmath.NewInt(3),
					Remainders:      sdk.NewDecCoins(),
				}
				rewards = sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
				)
			},
			types.Vault{
				Key:          VaultKeyWithRewards,
				VaultAddress: VaultWithRewardsAddress.String(),
				IsActive:     true,
				RewardsPerPower: sdk.NewDecCoins(
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
			suite.SetupTest()
			ctx = suite.ctx
			testCase.malleate()

			suite.restakeKeeper.SetVault(ctx, vault)
			err = suite.restakeKeeper.AddRewards(
				ctx,
				RewarderAddress,
				vault.Key,
				rewards,
			)
			suite.Require().NoError(err)

			vault, found := suite.restakeKeeper.GetVault(ctx, vault.Key)
			suite.Require().True(found)
			suite.Require().Equal(testCase.expVault, vault)
		})
	}
}

func (suite *KeeperTestSuite) TestIsActiveVault() {
	ctx := suite.ctx
	suite.setupState()

	// case - valid vault
	for _, expVault := range suite.validVaults {
		isActive := suite.restakeKeeper.IsActiveVault(ctx, expVault.Key)
		suite.Require().Equal(expVault.IsActive, isActive)
	}

	// case - no vault
	isActive := suite.restakeKeeper.IsActiveVault(ctx, InvalidVaultKey)
	suite.Require().Equal(false, isActive)
}

func (suite *KeeperTestSuite) TestDeactivateVault() {
	ctx := suite.ctx
	suite.setupState()

	// error case -  no vault
	err := suite.restakeKeeper.DeactivateVault(ctx, InvalidVaultKey)
	suite.Require().Error(err)

	// error case - vault is deactivated
	err = suite.restakeKeeper.DeactivateVault(ctx, InactiveVaultKey)
	suite.Require().Error(err)

	// success case
	err = suite.restakeKeeper.DeactivateVault(ctx, VaultKeyWithRewards)
	suite.Require().NoError(err)
	vault, found := suite.restakeKeeper.GetVault(ctx, VaultKeyWithRewards)
	suite.Require().True(found)
	suite.Require().Equal(false, vault.IsActive)
}

func (suite *KeeperTestSuite) TestGetSetVault() {
	ctx := suite.ctx

	// set
	expectedVaults := suite.validVaults
	for _, expVault := range expectedVaults {
		suite.restakeKeeper.SetVault(ctx, expVault)

		// get
		vault, found := suite.restakeKeeper.GetVault(ctx, expVault.Key)
		suite.Require().True(found)
		suite.Require().Equal(expVault, vault)

		// must get
		vault = suite.restakeKeeper.MustGetVault(ctx, expVault.Key)
		suite.Require().Equal(expVault, vault)
	}

	// get
	vaults := suite.restakeKeeper.GetVaults(ctx)
	suite.Require().Equal(expectedVaults, vaults)

	_, found := suite.restakeKeeper.GetVault(ctx, "nonVault")
	suite.Require().False(found)

	// must get
	suite.Require().Panics(func() {
		_ = suite.restakeKeeper.MustGetVault(ctx, "nonVault")
	})
}
