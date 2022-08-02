package keeper

import (
	"errors"
	"io"

	snapshot "github.com/cosmos/cosmos-sdk/snapshots/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	protoio "github.com/gogo/protobuf/io"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/bandprotocol/chain/v2/pkg/filecache"
	"github.com/bandprotocol/chain/v2/pkg/gzip"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

var _ snapshot.ExtensionSnapshotter = &OracleSnapshotter{}

// SnapshotFormat format 1 is just gzipped byte code for each oracle scripts and data sources.
const SnapshotFormat = 1

type OracleSnapshotter struct {
	keeper *Keeper
	cms    sdk.MultiStore
}

func NewOracleSnapshotter(cms sdk.MultiStore, keeper *Keeper) *OracleSnapshotter {
	return &OracleSnapshotter{
		keeper: keeper,
		cms:    cms,
	}
}

func (os *OracleSnapshotter) SnapshotName() string {
	return types.ModuleName
}

func (os *OracleSnapshotter) SnapshotFormat() uint32 {
	return SnapshotFormat
}

func (os *OracleSnapshotter) SupportedFormats() []uint32 {
	return []uint32{SnapshotFormat}
}

func (os *OracleSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	cacheMS, err := os.cms.CacheMultiStoreWithVersion(int64(height))
	if err != nil {
		return err
	}

	ctx := sdk.NewContext(cacheMS, tmproto.Header{}, false, log.NewNopLogger())
	seenBefore := make(map[string]bool)

	// write all oracle scripts to snapshot
	oracleScripts := os.keeper.GetAllOracleScripts(ctx)
	for _, oracleScript := range oracleScripts {
		if err := writeFileToSnapshot(protoWriter, oracleScript.Filename, os.keeper, seenBefore); err != nil {
			return err
		}
	}

	// write all data sources to snapshot
	dataSources := os.keeper.GetAllDataSources(ctx)
	for _, dataSource := range dataSources {
		if err := writeFileToSnapshot(protoWriter, dataSource.Filename, os.keeper, seenBefore); err != nil {
			return err
		}
	}

	return nil
}

func (os *OracleSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshot.SnapshotItem, error) {
	if format == SnapshotFormat {
		return os.processAllItems(height, protoReader, restoreV1, finalizeV1)
	}
	return snapshot.SnapshotItem{}, snapshot.ErrUnknownFormat
}

func (os *OracleSnapshotter) processAllItems(
	height uint64,
	protoReader protoio.Reader,
	restore func(sdk.Context, *Keeper, []byte, map[string]bool) error,
	finalize func(sdk.Context, *Keeper, map[string]bool) error,
) (snapshot.SnapshotItem, error) {
	ctx := sdk.NewContext(os.cms, tmproto.Header{Height: int64(height)}, false, log.NewNopLogger())

	// get all filename that we need to find and construct a map to store found status
	foundCode := make(map[string]bool)
	oracleScripts := os.keeper.GetAllOracleScripts(ctx)
	for _, oracleScript := range oracleScripts {
		foundCode[oracleScript.Filename] = false
	}
	dataSources := os.keeper.GetAllDataSources(ctx)
	for _, dataSource := range dataSources {
		foundCode[dataSource.Filename] = false
	}

	// keep the last snapshot item that is not for our module and return it for snapshot manager to handle.
	var item snapshot.SnapshotItem
	for {
		item = snapshot.SnapshotItem{}
		err := protoReader.ReadMsg(&item)
		if err == io.EOF {
			break
		} else if err != nil {
			return snapshot.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}

		payload := item.GetExtensionPayload()
		if payload == nil {
			break
		}

		if err := restore(ctx, os.keeper, payload.Payload, foundCode); err != nil {
			return snapshot.SnapshotItem{}, sdkerrors.Wrap(err, "processing snapshot item")
		}
	}

	return item, finalize(ctx, os.keeper, foundCode)
}

func writeFileToSnapshot(
	protoWriter protoio.Writer,
	filename string,
	k *Keeper,
	seenBefore map[string]bool,
) error {
	// no need to write if we write it before
	if seenBefore[filename] {
		return nil
	}
	seenBefore[filename] = true

	// get byte code from filename
	bytes, err := k.fileCache.GetFile(filename)
	if err != nil {
		return err
	}

	// zip it
	compressBytes, err := gzip.Compress(bytes)
	if err != nil {
		return err
	}

	// write it to snapshot
	if err = snapshot.WriteExtensionItem(protoWriter, compressBytes); err != nil {
		return err
	}

	return nil
}

func restoreV1(ctx sdk.Context, k *Keeper, compressedCode []byte, foundCode map[string]bool) error {
	// uncompress code
	code, err := gzip.Uncompress(
		compressedCode,
		max(types.MaxExecutableSize, types.MaxWasmCodeSize, types.MaxCompiledWasmCodeSize),
	)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
	}

	// check if we really need this file or not first
	filename := filecache.GetFilename(code)
	found, required := foundCode[filename]

	if !required {
		return errors.New("found unexpected code in the snapshot")
	}

	if !found {
		// add the file to disk
		foundCode[filename] = true
		k.fileCache.AddFile(code)
	}

	return nil
}

func finalizeV1(ctx sdk.Context, k *Keeper, foundCode map[string]bool) error {
	// check if there is any required code that we can't find in restore process
	for _, found := range foundCode {
		if !found {
			return errors.New("some code is missing from the snapshot")
		}
	}
	return nil
}

func max(arr ...int64) int64 {
	var maximum int64 = 0
	for _, value := range arr {
		if value > maximum {
			maximum = value
		}
	}

	return maximum
}
