package keeper_test

import (
	"math/big"

	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

// CollectFee returns an error for very large data source fees.
func (suite *KeeperTestSuite) TestCollectFeeLargeAmount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	huge := sdk.NewCoins(sdk.NewCoin("uband", math.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 254))))
	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{huge})

	_, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, huge, 4, raws)
	require.ErrorIs(err, types.ErrNotEnoughFee)

	// Two data sources with very large fees.
	maxFee := sdk.NewCoins(sdk.NewCoin("uband", math.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))))
	raws = rawRequestsFromFees(ctx, k, []sdk.Coins{maxFee, maxFee})
	suite.bankKeeper.EXPECT().SendCoins(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	_, err = k.CollectFee(ctx, bandtesting.FeePayer.Address, maxFee, 1, raws)
	require.ErrorIs(err, types.ErrNotEnoughFee)
}

// Data source treasury must be a regular account.
func (suite *KeeperTestSuite) TestDataSourceTreasuryValidation() {
	ctx := suite.ctx
	require := suite.Require()
	msgServer := keeper.NewMsgServerImpl(suite.oracleKeeper)

	bondedPool := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	suite.bankKeeper.EXPECT().BlockedAddr(bondedPool).Return(true).AnyTimes()
	suite.bankKeeper.EXPECT().BlockedAddr(bandtesting.Treasury.Address).Return(false).AnyTimes()

	exec := []byte("executable")

	_, err := msgServer.CreateDataSource(ctx, types.NewMsgCreateDataSource(
		"name", "desc", exec, bandtesting.Coins1band, bondedPool, bandtesting.Owner.Address, bandtesting.Owner.Address,
	))
	require.ErrorIs(err, types.ErrInvalidTreasury)

	_, err = msgServer.CreateDataSource(ctx, types.NewMsgCreateDataSource(
		"name", "desc", exec, bandtesting.Coins1band, bandtesting.Treasury.Address, bandtesting.Owner.Address, bandtesting.Owner.Address,
	))
	require.NoError(err)

	id := types.DataSourceID(suite.oracleKeeper.GetDataSourceCount(ctx))
	_, err = msgServer.EditDataSource(ctx, types.NewMsgEditDataSource(
		id, "name", "desc", exec, bandtesting.Coins1band, bondedPool, bandtesting.Owner.Address, bandtesting.Owner.Address,
	))
	require.ErrorIs(err, types.ErrInvalidTreasury)
	require.Equal(bandtesting.Treasury.Address.String(), suite.oracleKeeper.MustGetDataSource(ctx, id).Treasury)
}
