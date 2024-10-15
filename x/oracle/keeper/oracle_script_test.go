package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) TestHasOracleScript() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We should not have a oracle script ID 42 without setting it.
	require.False(k.HasOracleScript(ctx, 42))
	// After we set it, we should be able to find it.
	k.SetOracleScript(ctx, 42, types.NewOracleScript(
		bandtesting.Owner.Address, basicName, basicDesc, basicFilename, basicSchema, basicSourceCodeURL,
	))
	require.True(k.HasOracleScript(ctx, 42))
}

func (suite *KeeperTestSuite) TestSetterGetterOracleScript() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Getting a non-existent oracle script should return error.
	_, err := k.GetOracleScript(ctx, 42)
	require.ErrorIs(err, types.ErrOracleScriptNotFound)
	require.Panics(func() { _ = k.MustGetOracleScript(ctx, 42) })
	// Creates some basic oracle scripts.
	oracleScript1 := types.NewOracleScript(
		bandtesting.Alice.Address, "NAME1", "DESCRIPTION1", "FILENAME1", basicSchema, basicSourceCodeURL,
	)
	oracleScript2 := types.NewOracleScript(
		bandtesting.Bob.Address, "NAME2", "DESCRIPTION2", "FILENAME2", basicSchema, basicSourceCodeURL,
	)
	// Sets id 42 with oracle script 1 and id 42 with oracle script 2.
	k.SetOracleScript(ctx, 42, oracleScript1)
	k.SetOracleScript(ctx, 43, oracleScript2)
	// Checks that Get and MustGet perform correctly.
	oracleScript1Res, err := k.GetOracleScript(ctx, 42)
	require.Nil(err)
	require.Equal(oracleScript1, oracleScript1Res)
	require.Equal(oracleScript1, k.MustGetOracleScript(ctx, 42))
	oracleScript2Res, err := k.GetOracleScript(ctx, 43)
	require.Nil(err)
	require.Equal(oracleScript2, oracleScript2Res)
	require.Equal(oracleScript2, k.MustGetOracleScript(ctx, 43))
	// Replaces id 42 with another oracle script.
	k.SetOracleScript(ctx, 42, oracleScript2)
	require.NotEqual(oracleScript1, k.MustGetOracleScript(ctx, 42))
	require.Equal(oracleScript2, k.MustGetOracleScript(ctx, 42))
}

func (suite *KeeperTestSuite) TestAddEditOracleScriptBasic() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Creates some basic oracle scripts.
	oracleScript1 := types.NewOracleScript(
		bandtesting.Alice.Address, "NAME1", "DESCRIPTION1", "FILENAME1", basicSchema, basicSourceCodeURL,
	)
	oracleScript2 := types.NewOracleScript(
		bandtesting.Bob.Address, "NAME2", "DESCRIPTION2", "FILENAME2", basicSchema, basicSourceCodeURL,
	)
	// Adds a new oracle script to the store. We should be able to retrieve it back.
	id := k.AddOracleScript(ctx, oracleScript1)
	require.Equal(oracleScript1, k.MustGetOracleScript(ctx, id))
	require.NotEqual(oracleScript2, k.MustGetOracleScript(ctx, id))
	// Edits the oracle script. We should get the updated oracle script.
	owner, err := sdk.AccAddressFromBech32(oracleScript2.Owner)
	require.NoError(err)
	require.NotPanics(func() {
		k.MustEditOracleScript(ctx, id, types.NewOracleScript(
			owner, oracleScript2.Name, oracleScript2.Description, oracleScript2.Filename,
			oracleScript2.Schema, oracleScript2.SourceCodeURL,
		))
	})
	require.NotEqual(oracleScript1, k.MustGetOracleScript(ctx, id))
	require.Equal(oracleScript2, k.MustGetOracleScript(ctx, id))
}

func (suite *KeeperTestSuite) TestAddEditOracleScriptDoNotModify() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Creates some basic oracle scripts.
	oracleScript1 := types.NewOracleScript(
		bandtesting.Alice.Address, "NAME1", "DESCRIPTION1", "FILENAME1", basicSchema, basicSourceCodeURL,
	)
	oracleScript2 := types.NewOracleScript(
		bandtesting.Bob.Address, types.DoNotModify, types.DoNotModify, "FILENAME2",
		types.DoNotModify, types.DoNotModify,
	)
	// Adds a new oracle script to the store. We should be able to retrieve it back.
	id := k.AddOracleScript(ctx, oracleScript1)
	require.Equal(oracleScript1, k.MustGetOracleScript(ctx, id))
	require.NotEqual(oracleScript2, k.MustGetOracleScript(ctx, id))
	// Edits the oracle script. We should get the updated oracle script.
	require.NotPanics(func() { k.MustEditOracleScript(ctx, id, oracleScript2) })
	oracleScriptRes := k.MustGetOracleScript(ctx, id)
	require.NotEqual(oracleScriptRes, oracleScript1)
	require.NotEqual(oracleScriptRes, oracleScript2)
	require.Equal(oracleScriptRes.Owner, oracleScript2.Owner)
	require.Equal(oracleScriptRes.Name, oracleScript1.Name)
	require.Equal(oracleScriptRes.Description, oracleScript1.Description)
	require.Equal(oracleScriptRes.Filename, oracleScript2.Filename)
	require.Equal(oracleScriptRes.Schema, oracleScript1.Schema)
	require.Equal(oracleScriptRes.SourceCodeURL, oracleScript1.SourceCodeURL)
}

func (suite *KeeperTestSuite) TestAddOracleScriptMustReturnCorrectID() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	genesisCount := k.GetOracleScriptCount(ctx)
	// Every new oracle script we add should return a new ID.
	id1 := k.AddOracleScript(ctx, types.NewOracleScript(
		bandtesting.Owner.Address, basicName, basicDesc, basicFilename, basicSchema, basicSourceCodeURL,
	))
	require.Equal(types.OracleScriptID(genesisCount+1), id1)
	// Adds another oracle script so now ID should increase by 2.
	id2 := k.AddOracleScript(ctx, types.NewOracleScript(
		bandtesting.Owner.Address, basicName, basicDesc, basicFilename, basicSchema, basicSourceCodeURL,
	))
	require.Equal(types.OracleScriptID(genesisCount+2), id2)
	// Finally we expect the oracle script to increase as well.
	require.Equal(genesisCount+2, k.GetOracleScriptCount(ctx))
}

func (suite *KeeperTestSuite) TestEditNonExistentOracleScript() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Editing a non-existent oracle script should return error.
	require.Panics(func() {
		k.MustEditOracleScript(ctx, 42, types.NewOracleScript(
			bandtesting.Owner.Address, basicName, basicDesc, basicFilename, basicSchema, basicSourceCodeURL,
		))
	})
}

func (suite *KeeperTestSuite) TestGetAllOracleScripts() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	oracleScripts := bandtesting.GenerateOracleScripts(suite.homeDir)

	// We should be able to get all genesis oracle scripts.
	require.Equal(
		oracleScripts,
		k.GetAllOracleScripts(ctx),
	)
}

func (suite *KeeperTestSuite) TestAddOracleScriptFile() {
	k := suite.oracleKeeper
	require := suite.Require()

	// Code should be perfectly compilable.
	compiledCode, err := bandtesting.OwasmVM.Compile(testdata.WasmExtra1, types.MaxCompiledWasmCodeSize)
	require.NoError(err)
	// We start by adding the Owasm content to the storage.
	filename, err := k.AddOracleScriptFile(testdata.WasmExtra1)
	require.NoError(err)
	// If we get by file name, we should get the compiled content back.
	require.Equal(compiledCode, k.GetFile(filename))
	// If we try to add do-not-modify, we should just get do-not-modify back.
	filename, err = k.AddOracleScriptFile(types.DoNotModifyBytes)
	require.NoError(err)
	require.Equal(types.DoNotModify, filename)
	// We should not be able to add a non-wasm file.
	_, err = k.AddOracleScriptFile([]byte("code"))
	require.ErrorIs(err, types.ErrOwasmCompilation)
}
