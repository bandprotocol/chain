package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

var govAuthority = authtypes.NewModuleAddress(govtypes.ModuleName)

func (suite *KeeperTestSuite) setupMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(suite.oracleKeeper)
}

func (suite *KeeperTestSuite) TestMsgCreateDataSourceRequiresAuthority() {
	msgServer := suite.setupMsgServer()
	count := suite.oracleKeeper.GetDataSourceCount(suite.ctx)

	// a regular account, even the intended owner, cannot create a data source
	msg := types.NewMsgCreateDataSource(
		basicName, basicDesc, []byte("executable"), bandtesting.EmptyCoins, treasury, owner, owner,
	)
	_, err := msgServer.CreateDataSource(suite.ctx, msg)
	suite.Require().ErrorIs(err, govtypes.ErrInvalidSigner)
	suite.Require().Equal(count, suite.oracleKeeper.GetDataSourceCount(suite.ctx))

	// the governance module account can
	msg = types.NewMsgCreateDataSource(
		basicName, basicDesc, []byte("executable"), bandtesting.EmptyCoins, treasury, owner, govAuthority,
	)
	_, err = msgServer.CreateDataSource(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().Equal(count+1, suite.oracleKeeper.GetDataSourceCount(suite.ctx))

	ds, err := suite.oracleKeeper.GetDataSource(suite.ctx, types.DataSourceID(count+1))
	suite.Require().NoError(err)
	suite.Require().Equal(owner.String(), ds.Owner)
	suite.Require().Equal(basicName, ds.Name)
}

func (suite *KeeperTestSuite) TestMsgEditDataSourceRequiresAuthority() {
	msgServer := suite.setupMsgServer()

	id := suite.oracleKeeper.AddDataSource(suite.ctx, types.NewDataSource(
		owner, basicName, basicDesc, basicFilename, bandtesting.EmptyCoins, treasury,
	))

	// the current owner can no longer edit directly
	msg := types.NewMsgEditDataSource(
		id, "NEW_NAME", basicDesc, []byte("executable"), bandtesting.EmptyCoins, treasury, owner, owner,
	)
	_, err := msgServer.EditDataSource(suite.ctx, msg)
	suite.Require().ErrorIs(err, govtypes.ErrInvalidSigner)

	ds, err := suite.oracleKeeper.GetDataSource(suite.ctx, id)
	suite.Require().NoError(err)
	suite.Require().Equal(basicName, ds.Name)

	// governance can edit and reassign the owner
	msg = types.NewMsgEditDataSource(
		id, "NEW_NAME", basicDesc, []byte("executable"), bandtesting.EmptyCoins, treasury, alice, govAuthority,
	)
	_, err = msgServer.EditDataSource(suite.ctx, msg)
	suite.Require().NoError(err)

	ds, err = suite.oracleKeeper.GetDataSource(suite.ctx, id)
	suite.Require().NoError(err)
	suite.Require().Equal("NEW_NAME", ds.Name)
	suite.Require().Equal(alice.String(), ds.Owner)

	// editing a non-existent data source still fails, even for governance
	msg = types.NewMsgEditDataSource(
		id+1, "NEW_NAME", basicDesc, []byte("executable"), bandtesting.EmptyCoins, treasury, alice, govAuthority,
	)
	_, err = msgServer.EditDataSource(suite.ctx, msg)
	suite.Require().ErrorIs(err, types.ErrDataSourceNotFound)
}

func (suite *KeeperTestSuite) TestMsgCreateOracleScriptRequiresAuthority() {
	msgServer := suite.setupMsgServer()
	count := suite.oracleKeeper.GetOracleScriptCount(suite.ctx)

	msg := types.NewMsgCreateOracleScript(
		basicName, basicDesc, basicSchema, basicSourceCodeURL, testdata.Wasm1, owner, owner,
	)
	_, err := msgServer.CreateOracleScript(suite.ctx, msg)
	suite.Require().ErrorIs(err, govtypes.ErrInvalidSigner)
	suite.Require().Equal(count, suite.oracleKeeper.GetOracleScriptCount(suite.ctx))

	msg = types.NewMsgCreateOracleScript(
		basicName, basicDesc, basicSchema, basicSourceCodeURL, testdata.Wasm1, owner, govAuthority,
	)
	_, err = msgServer.CreateOracleScript(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().Equal(count+1, suite.oracleKeeper.GetOracleScriptCount(suite.ctx))

	os, err := suite.oracleKeeper.GetOracleScript(suite.ctx, types.OracleScriptID(count+1))
	suite.Require().NoError(err)
	suite.Require().Equal(owner.String(), os.Owner)
	suite.Require().Equal(basicName, os.Name)
}

func (suite *KeeperTestSuite) TestMsgEditOracleScriptRequiresAuthority() {
	msgServer := suite.setupMsgServer()

	id := suite.oracleKeeper.AddOracleScript(suite.ctx, types.NewOracleScript(
		owner, basicName, basicDesc, basicFilename, basicSchema, basicSourceCodeURL,
	))

	// the current owner can no longer edit directly
	msg := types.NewMsgEditOracleScript(
		id, "NEW_NAME", basicDesc, basicSchema, basicSourceCodeURL, testdata.Wasm1, owner, owner,
	)
	_, err := msgServer.EditOracleScript(suite.ctx, msg)
	suite.Require().ErrorIs(err, govtypes.ErrInvalidSigner)

	os, err := suite.oracleKeeper.GetOracleScript(suite.ctx, id)
	suite.Require().NoError(err)
	suite.Require().Equal(basicName, os.Name)

	// governance can edit and reassign the owner
	msg = types.NewMsgEditOracleScript(
		id, "NEW_NAME", basicDesc, basicSchema, basicSourceCodeURL, testdata.Wasm1, alice, govAuthority,
	)
	_, err = msgServer.EditOracleScript(suite.ctx, msg)
	suite.Require().NoError(err)

	os, err = suite.oracleKeeper.GetOracleScript(suite.ctx, id)
	suite.Require().NoError(err)
	suite.Require().Equal("NEW_NAME", os.Name)
	suite.Require().Equal(alice.String(), os.Owner)

	// editing a non-existent oracle script still fails, even for governance
	msg = types.NewMsgEditOracleScript(
		id+1, "NEW_NAME", basicDesc, basicSchema, basicSourceCodeURL, testdata.Wasm1, alice, govAuthority,
	)
	_, err = msgServer.EditOracleScript(suite.ctx, msg)
	suite.Require().ErrorIs(err, types.ErrOracleScriptNotFound)
}

func (suite *KeeperTestSuite) TestMsgUpdateParamsRequiresAuthority() {
	msgServer := suite.setupMsgServer()

	params := types.DefaultParams()
	_, err := msgServer.UpdateParams(suite.ctx, types.NewMsgUpdateParams(alice.String(), params))
	suite.Require().ErrorIs(err, govtypes.ErrInvalidSigner)

	_, err = msgServer.UpdateParams(suite.ctx, types.NewMsgUpdateParams(govAuthority.String(), params))
	suite.Require().NoError(err)
}
