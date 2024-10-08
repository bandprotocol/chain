package keeper_test

// TODO: Fix tests
// import (
// 	"testing"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/require"

// 	bandtesting "github.com/bandprotocol/chain/v3/testing"
// 	"github.com/bandprotocol/chain/v3/testing/testdata"
// 	"github.com/bandprotocol/chain/v3/x/oracle/types"
// )

// func TestHasOracleScript(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// We should not have a oracle script ID 42 without setting it.
// 	require.False(t, k.HasOracleScript(ctx, 42))
// 	// After we set it, we should be able to find it.
// 	k.SetOracleScript(ctx, 42, types.NewOracleScript(
// 		bandtesting.Owner.Address, BasicName, BasicDesc, BasicFilename, BasicSchema, BasicSourceCodeURL,
// 	))
// 	require.True(t, k.HasOracleScript(ctx, 42))
// }

// func TestSetterGetterOracleScript(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// Getting a non-existent oracle script should return error.
// 	_, err := k.GetOracleScript(ctx, 42)
// 	require.ErrorIs(t, err, types.ErrOracleScriptNotFound)
// 	require.Panics(t, func() { _ = k.MustGetOracleScript(ctx, 42) })
// 	// Creates some basic oracle scripts.
// 	oracleScript1 := types.NewOracleScript(
// 		bandtesting.Alice.Address, "NAME1", "DESCRIPTION1", "FILENAME1", BasicSchema, BasicSourceCodeURL,
// 	)
// 	oracleScript2 := types.NewOracleScript(
// 		bandtesting.Bob.Address, "NAME2", "DESCRIPTION2", "FILENAME2", BasicSchema, BasicSourceCodeURL,
// 	)
// 	// Sets id 42 with oracle script 1 and id 42 with oracle script 2.
// 	k.SetOracleScript(ctx, 42, oracleScript1)
// 	k.SetOracleScript(ctx, 43, oracleScript2)
// 	// Checks that Get and MustGet perform correctly.
// 	oracleScript1Res, err := k.GetOracleScript(ctx, 42)
// 	require.Nil(t, err)
// 	require.Equal(t, oracleScript1, oracleScript1Res)
// 	require.Equal(t, oracleScript1, k.MustGetOracleScript(ctx, 42))
// 	oracleScript2Res, err := k.GetOracleScript(ctx, 43)
// 	require.Nil(t, err)
// 	require.Equal(t, oracleScript2, oracleScript2Res)
// 	require.Equal(t, oracleScript2, k.MustGetOracleScript(ctx, 43))
// 	// Replaces id 42 with another oracle script.
// 	k.SetOracleScript(ctx, 42, oracleScript2)
// 	require.NotEqual(t, oracleScript1, k.MustGetOracleScript(ctx, 42))
// 	require.Equal(t, oracleScript2, k.MustGetOracleScript(ctx, 42))
// }

// func TestAddEditOracleScriptBasic(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// Creates some basic oracle scripts.
// 	oracleScript1 := types.NewOracleScript(
// 		bandtesting.Alice.Address, "NAME1", "DESCRIPTION1", "FILENAME1", BasicSchema, BasicSourceCodeURL,
// 	)
// 	oracleScript2 := types.NewOracleScript(
// 		bandtesting.Bob.Address, "NAME2", "DESCRIPTION2", "FILENAME2", BasicSchema, BasicSourceCodeURL,
// 	)
// 	// Adds a new oracle script to the store. We should be able to retrieve it back.
// 	id := k.AddOracleScript(ctx, oracleScript1)
// 	require.Equal(t, oracleScript1, k.MustGetOracleScript(ctx, id))
// 	require.NotEqual(t, oracleScript2, k.MustGetOracleScript(ctx, id))
// 	// Edits the oracle script. We should get the updated oracle script.
// 	owner, err := sdk.AccAddressFromBech32(oracleScript2.Owner)
// 	require.NoError(t, err)
// 	require.NotPanics(t, func() {
// 		k.MustEditOracleScript(ctx, id, types.NewOracleScript(
// 			owner, oracleScript2.Name, oracleScript2.Description, oracleScript2.Filename,
// 			oracleScript2.Schema, oracleScript2.SourceCodeURL,
// 		))
// 	})
// 	require.NotEqual(t, oracleScript1, k.MustGetOracleScript(ctx, id))
// 	require.Equal(t, oracleScript2, k.MustGetOracleScript(ctx, id))
// }

// func TestAddEditOracleScriptDoNotModify(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// Creates some basic oracle scripts.
// 	oracleScript1 := types.NewOracleScript(
// 		bandtesting.Alice.Address, "NAME1", "DESCRIPTION1", "FILENAME1", BasicSchema, BasicSourceCodeURL,
// 	)
// 	oracleScript2 := types.NewOracleScript(
// 		bandtesting.Bob.Address, types.DoNotModify, types.DoNotModify, "FILENAME2",
// 		types.DoNotModify, types.DoNotModify,
// 	)
// 	// Adds a new oracle script to the store. We should be able to retrieve it back.
// 	id := k.AddOracleScript(ctx, oracleScript1)
// 	require.Equal(t, oracleScript1, k.MustGetOracleScript(ctx, id))
// 	require.NotEqual(t, oracleScript2, k.MustGetOracleScript(ctx, id))
// 	// Edits the oracle script. We should get the updated oracle script.
// 	require.NotPanics(t, func() { k.MustEditOracleScript(ctx, id, oracleScript2) })
// 	oracleScriptRes := k.MustGetOracleScript(ctx, id)
// 	require.NotEqual(t, oracleScriptRes, oracleScript1)
// 	require.NotEqual(t, oracleScriptRes, oracleScript2)
// 	require.Equal(t, oracleScriptRes.Owner, oracleScript2.Owner)
// 	require.Equal(t, oracleScriptRes.Name, oracleScript1.Name)
// 	require.Equal(t, oracleScriptRes.Description, oracleScript1.Description)
// 	require.Equal(t, oracleScriptRes.Filename, oracleScript2.Filename)
// 	require.Equal(t, oracleScriptRes.Schema, oracleScript1.Schema)
// 	require.Equal(t, oracleScriptRes.SourceCodeURL, oracleScript1.SourceCodeURL)
// }

// func TestAddOracleScriptMustReturnCorrectID(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// Initially we expect the oracle script count to be what we have on genesis state.
// 	genesisCount := uint64(len(bandtesting.OracleScripts)) - 1
// 	require.Equal(t, genesisCount, k.GetOracleScriptCount(ctx))
// 	// Every new oracle script we add should return a new ID.
// 	id1 := k.AddOracleScript(ctx, types.NewOracleScript(
// 		bandtesting.Owner.Address, BasicName, BasicDesc, BasicFilename, BasicSchema, BasicSourceCodeURL,
// 	))
// 	require.Equal(t, types.OracleScriptID(genesisCount+1), id1)
// 	// Adds another oracle script so now ID should increase by 2.
// 	id2 := k.AddOracleScript(ctx, types.NewOracleScript(
// 		bandtesting.Owner.Address, BasicName, BasicDesc, BasicFilename, BasicSchema, BasicSourceCodeURL,
// 	))
// 	require.Equal(t, types.OracleScriptID(genesisCount+2), id2)
// 	// Finally we expect the oracle script to increase as well.
// 	require.Equal(t, genesisCount+2, k.GetOracleScriptCount(ctx))
// }

// func TestEditNonExistentOracleScript(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// Editing a non-existent oracle script should return error.
// 	require.Panics(t, func() {
// 		k.MustEditOracleScript(ctx, 42, types.NewOracleScript(
// 			bandtesting.Owner.Address, BasicName, BasicDesc, BasicFilename, BasicSchema, BasicSourceCodeURL,
// 		))
// 	})
// }

// func TestGetAllOracleScripts(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// We should be able to get all genesis oracle scripts.
// 	require.Equal(t, bandtesting.OracleScripts[1:], k.GetAllOracleScripts(ctx))
// }

// func TestAddOracleScriptFile(t *testing.T) {
// 	app, _ := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// Code should be perfectly compilable.
// 	compiledCode, err := bandtesting.OwasmVM.Compile(testdata.WasmExtra1, types.MaxCompiledWasmCodeSize)
// 	require.NoError(t, err)
// 	// We start by adding the Owasm content to the storage.
// 	filename, err := k.AddOracleScriptFile(testdata.WasmExtra1)
// 	require.NoError(t, err)
// 	// If we get by file name, we should get the compiled content back.
// 	require.Equal(t, compiledCode, k.GetFile(filename))
// 	// If we try to add do-not-modify, we should just get do-not-modify back.
// 	filename, err = k.AddOracleScriptFile(types.DoNotModifyBytes)
// 	require.NoError(t, err)
// 	require.Equal(t, types.DoNotModify, filename)
// 	// We should not be able to add a non-wasm file.
// 	_, err = k.AddOracleScriptFile([]byte("code"))
// 	require.ErrorIs(t, err, types.ErrOwasmCompilation)
// }
