package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

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
