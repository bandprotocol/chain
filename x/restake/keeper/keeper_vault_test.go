package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/restake/types"
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
		Key:      "newVault",
		IsActive: true,
	}, vault)
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
	suite.Require().False(isActive)
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
	err = suite.restakeKeeper.DeactivateVault(ctx, ActiveVaultKey)
	suite.Require().NoError(err)
	vault, found := suite.restakeKeeper.GetVault(ctx, ActiveVaultKey)
	suite.Require().True(found)
	suite.Require().False(vault.IsActive)
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
