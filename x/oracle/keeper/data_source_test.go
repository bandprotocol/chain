package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) TestHasDataSource() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We should not have a data source ID 42 without setting it.
	require.False(k.HasDataSource(ctx, 42))
	// After we set it, we should be able to find it.
	k.SetDataSource(ctx, 42, types.NewDataSource(
		owner,
		basicName,
		basicDesc,
		basicFilename,
		coinsZero,
		types.KeyExpirationBlockCount,
	))
	require.True(k.HasDataSource(ctx, 42))
}

func (suite *KeeperTestSuite) TestSetterGetterDataSource() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Getting a non-existent data source should return error.
	_, err := k.GetDataSource(ctx, 42)
	require.ErrorIs(err, types.ErrDataSourceNotFound)
	require.Panics(func() { _ = k.MustGetDataSource(ctx, 42) })
	// Creates some basic data sources.
	dataSource1 := types.NewDataSource(
		alice,
		"NAME1",
		"DESCRIPTION1",
		"filename1",
		emptyCoins,
		treasury,
	)
	dataSource2 := types.NewDataSource(
		bob,
		"NAME2",
		"DESCRIPTION2",
		"filename2",
		emptyCoins,
		treasury,
	)
	// Sets id 42 with data soure 1 and id 42 with data source 2.
	k.SetDataSource(ctx, 42, dataSource1)
	k.SetDataSource(ctx, 43, dataSource2)
	// Checks that Get and MustGet perform correctly.
	dataSource1Res, err := k.GetDataSource(ctx, 42)
	require.Nil(err)
	require.Equal(dataSource1, dataSource1Res)
	require.Equal(dataSource1, k.MustGetDataSource(ctx, 42))

	dataSource2Res, err := k.GetDataSource(ctx, 43)
	require.Nil(err)
	require.Equal(dataSource2, dataSource2Res)
	require.Equal(dataSource2, k.MustGetDataSource(ctx, 43))
	// Replaces id 42 with another data source.

	k.SetDataSource(ctx, 42, dataSource2)
	require.NotEqual(dataSource1, k.MustGetDataSource(ctx, 42))
	require.Equal(dataSource2, k.MustGetDataSource(ctx, 42))
}

// func (suite *KeeperTestSuite) TestAddDataSourceEditDataSourceBasic() {
// 	ctx := suite.ctx
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	// Creates some basic data sources.
// 	dataSource1 := types.NewDataSource(
// 		bandtesting.Alice.Address,
// 		"NAME1",
// 		"DESCRIPTION1",
// 		"FILENAME1",
// 		bandtesting.EmptyCoins,
// 		bandtesting.Treasury.Address,
// 	)
// 	dataSource2 := types.NewDataSource(
// 		bandtesting.Bob.Address,
// 		"NAME2",
// 		"DESCRIPTION2",
// 		"FILENAME2",
// 		bandtesting.EmptyCoins,
// 		bandtesting.Treasury.Address,
// 	)
// 	// Adds a new data source to the store. We should be able to retrieve it back.
// 	id := k.AddDataSource(ctx, dataSource1)
// 	require.Equal(t, dataSource1, k.MustGetDataSource(ctx, id))
// 	require.NotEqual(t, dataSource2, k.MustGetDataSource(ctx, id))
// 	owner, err := sdk.AccAddressFromBech32(dataSource2.Owner)
// 	require.NoError(t, err)
// 	treasury, err := sdk.AccAddressFromBech32(dataSource2.Treasury)
// 	require.NoError(t, err)
// 	// Edits the data source. We should get the updated data source.
// 	k.MustEditDataSource(ctx, id, types.NewDataSource(
// 		owner, dataSource2.Name, dataSource2.Description, dataSource2.Filename, bandtesting.EmptyCoins, treasury,
// 	))
// 	require.NotEqual(t, dataSource1, k.MustGetDataSource(ctx, id))
// 	require.Equal(t, dataSource2, k.MustGetDataSource(ctx, id))
// }

// func (suite *KeeperTestSuite) TestEditDataSourceDoNotModify() {
// 	ctx := suite.ctx
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	// Creates some basic data sources.
// 	dataSource1 := types.NewDataSource(
// 		bandtesting.Alice.Address,
// 		"NAME1",
// 		"DESCRIPTION1",
// 		"FILENAME1",
// 		bandtesting.EmptyCoins,
// 		bandtesting.Treasury.Address,
// 	)
// 	dataSource2 := types.NewDataSource(
// 		bandtesting.Bob.Address,
// 		types.DoNotModify,
// 		types.DoNotModify,
// 		"FILENAME2",
// 		bandtesting.EmptyCoins,
// 		bandtesting.Treasury.Address,
// 	)
// 	// Adds a new data source to the store. We should be able to retrieve it back.
// 	id := k.AddDataSource(ctx, dataSource1)
// 	require.Equal(t, dataSource1, k.MustGetDataSource(ctx, id))
// 	require.NotEqual(t, dataSource2, k.MustGetDataSource(ctx, id))
// 	// Edits the data source. We should get the updated data source.
// 	k.MustEditDataSource(ctx, id, dataSource2)
// 	dataSourceRes := k.MustGetDataSource(ctx, id)
// 	require.NotEqual(t, dataSourceRes, dataSource1)
// 	require.NotEqual(t, dataSourceRes, dataSource2)
// 	require.Equal(t, dataSourceRes.Owner, dataSource2.Owner)
// 	require.Equal(t, dataSourceRes.Name, dataSource1.Name)
// 	require.Equal(t, dataSourceRes.Description, dataSource1.Description)
// 	require.Equal(t, dataSourceRes.Filename, dataSource2.Filename)
// 	require.Equal(t, dataSourceRes.Fee, dataSource2.Fee)
// 	require.Equal(t, dataSourceRes.Treasury, dataSource2.Treasury)
// }

// func (suite *KeeperTestSuite) TestAddDataSourceDataSourceMustReturnCorrectID() {
// 	ctx := suite.ctx
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	// Initially we expect the data source count to be what we have on genesis state.
// 	genesisCount := uint64(len(bandtesting.DataSources)) - 1
// 	require.Equal(t, genesisCount, k.GetDataSourceCount(ctx))

// 	// Every new data source we add should return a new ID.
// 	id1 := k.AddDataSource(
// 		ctx,
// 		types.NewDataSource(
// 			bandtesting.Owner.Address,
// 			BasicName,
// 			BasicDesc,
// 			BasicFilename,
// 			bandtesting.EmptyCoins,
// 			bandtesting.Treasury.Address,
// 		),
// 	)
// 	require.Equal(t, types.DataSourceID(genesisCount+1), id1)
// 	// Adds another data source so now ID should increase by 2.
// 	id2 := k.AddDataSource(
// 		ctx,
// 		types.NewDataSource(
// 			bandtesting.Owner.Address,
// 			BasicName,
// 			BasicDesc,
// 			BasicFilename,
// 			bandtesting.EmptyCoins,
// 			bandtesting.Treasury.Address,
// 		),
// 	)
// 	require.Equal(t, types.DataSourceID(genesisCount+2), id2)
// 	// Finally we expect the data source to increase as well.
// 	require.Equal(t, genesisCount+2, k.GetDataSourceCount(ctx))
// }

// func (suite *KeeperTestSuite) TestEditDataSourceNonExistentDataSource() {
// 	ctx := suite.ctx
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	require.Panics(func() { k.MustEditDataSource(ctx, 9999, bandtesting.DataSources[1]) })
// }

// func (suite *KeeperTestSuite) TestGetAllDataSources() {
// 	ctx := suite.ctx
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	// We should be able to get all genesis data sources.
// 	require.Equal(bandtesting.DataSources[1:], k.GetAllDataSources(ctx))
// }

// func (suite *KeeperTestSuite) TestAddExecutableFile() {
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	// Adding do-not-modify should simply return do-not-modify.
// 	require.Equal(types.DoNotModify, k.AddExecutableFile(types.DoNotModifyBytes))
// 	// After we add an executable file, we should be able to retrieve it back.
// 	filename := k.AddExecutableFile([]byte("UNIQUE_EXEC_FOR_TestAddExecutableFile"))
// 	require.Equal([]byte("UNIQUE_EXEC_FOR_TestAddExecutableFile"), k.GetFile(filename))
// }
