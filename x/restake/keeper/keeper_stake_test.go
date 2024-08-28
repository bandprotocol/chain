package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestGetStakedPower() {
	ctx := suite.ctx
	suite.setupState()

	expPower, _ := sdkmath.NewIntFromString("50")
	power := suite.restakeKeeper.GetStakedPower(ctx, ValidAddress1)
	suite.Require().Equal(expPower, power)

	expPower, _ = sdkmath.NewIntFromString("0")
	power = suite.restakeKeeper.GetStakedPower(ctx, ValidAddress2)
	suite.Require().Equal(expPower, power)

	expPower, _ = sdkmath.NewIntFromString("10")
	power = suite.restakeKeeper.GetStakedPower(ctx, ValidAddress3)
	suite.Require().Equal(expPower, power)
}

func (suite *KeeperTestSuite) TestGetSetStake() {
	ctx := suite.ctx

	// set
	expectedStakes := suite.validStakes
	for _, expStake := range expectedStakes {
		// set
		acc := sdk.MustAccAddressFromBech32(expStake.StakerAddress)
		suite.restakeKeeper.SetStake(ctx, expStake)

		// get
		stake := suite.restakeKeeper.GetStake(ctx, acc)
		suite.Require().Equal(expStake, stake)
	}

	// get
	stakes := suite.restakeKeeper.GetStakes(ctx)
	suite.Require().Equal(expectedStakes, stakes)

	// get stake for valid address2 (no stake)
	stake := suite.restakeKeeper.GetStake(ctx, ValidAddress2)
	suite.Require().Equal(types.Stake{
		StakerAddress: ValidAddress2.String(),
		Coins:         sdk.Coins{},
	}, stake)

	// delete
	for _, expStake := range expectedStakes {
		acc := sdk.MustAccAddressFromBech32(expStake.StakerAddress)
		suite.restakeKeeper.DeleteStake(ctx, acc)

		// get
		stake := suite.restakeKeeper.GetStake(ctx, acc)
		suite.Require().Equal(types.Stake{
			StakerAddress: acc.String(),
			Coins:         sdk.Coins{},
		}, stake)
	}
}
