package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/restake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGetSetKey() {
	ctx := suite.ctx

	// set
	expectedKeys := []types.Key{
		{
			Name:            "Key0",
			PoolAddress:     "address0",
			IsActive:        true,
			TotalPower:      sdk.NewInt(0),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            "Key1",
			PoolAddress:     "address1",
			IsActive:        false,
			TotalPower:      sdk.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            "Key2",
			PoolAddress:     "address2",
			IsActive:        true,
			TotalPower:      sdk.NewInt(1000),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
	}
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
