package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
)

func TestSnapshotter(t *testing.T) {
	// setup source app
	srcDir := testutil.GetTempDir(t)
	srcApp := bandtesting.SetupWithCustomHome(false, srcDir)
	srcCtx := srcApp.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	srcKeeper := srcApp.OracleKeeper

	// create snapshot
	_, err := srcApp.Commit()
	require.NoError(t, err)
	srcHashToCode := getMappingHashToCode(srcCtx, &srcKeeper)
	snapshotHeight := uint64(srcApp.LastBlockHeight())
	snapshot, err := srcApp.SnapshotManager().Create(snapshotHeight)
	require.NoError(t, err)
	assert.NotNil(t, snapshot)

	// restore snapshot
	destDir := testutil.GetTempDir(t)
	destApp := bandtesting.SetupWithCustomHome(false, destDir)
	destCtx := destApp.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	destKeeper := destApp.OracleKeeper
	require.NoError(t, destApp.SnapshotManager().Restore(*snapshot))
	for i := uint32(0); i < snapshot.Chunks; i++ {
		chunkBz, err := srcApp.SnapshotManager().LoadChunk(snapshot.Height, snapshot.Format, i)
		require.NoError(t, err)
		end, err := destApp.SnapshotManager().RestoreChunk(chunkBz)
		require.NoError(t, err)
		if end {
			break
		}
	}
	destHashToCode := getMappingHashToCode(destCtx, &destKeeper)

	// compare src and dest
	assert.Equal(
		t,
		srcHashToCode,
		destHashToCode,
	)
}

func getMappingHashToCode(ctx sdk.Context, keeper *keeper.Keeper) map[string][]byte {
	hashToCode := make(map[string][]byte)
	oracleScripts := keeper.GetAllOracleScripts(ctx)
	for _, oracleScript := range oracleScripts {
		hashToCode[oracleScript.Filename] = keeper.GetFile(oracleScript.Filename)
	}
	dataSources := keeper.GetAllDataSources(ctx)
	for _, dataSource := range dataSources {
		hashToCode[dataSource.Filename] = keeper.GetFile(dataSource.Filename)
	}

	return hashToCode
}
