package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bandprotocol/chain/pkg/filecache"
	"github.com/bandprotocol/chain/x/oracle/types"
)

const (
	RollingSeedSizeInBytes = 32
)

type Keeper struct {
	storeKey         sdk.StoreKey
	cdc              codec.BinaryMarshaler
	fileCache        filecache.Cache
	feeCollectorName string
	paramstore       paramtypes.Subspace
	authKeeper       types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      types.DistrKeeper
	stakingKeeper    types.StakingKeeper
}

// NewKeeper creates a new oracle Keeper instance.
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, fileDir string, feeCollectorName string,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper, distrKeeper types.DistrKeeper, ps paramtypes.Subspace,
) Keeper {
	fmt.Print("!!->", ps)
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}
	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		fileCache:        filecache.New(fileDir),
		feeCollectorName: feeCollectorName,
		paramstore:       ps,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		distrKeeper:      distrKeeper,
		stakingKeeper:    stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParam returns the parameter as specified by key as an uint64.
func (k Keeper) GetParam(ctx sdk.Context, key []byte) (res uint64) {
	k.paramstore.Get(ctx, key, &res)
	return res
}

// SetParam saves the given key-value parameter to the store.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value uint64) {
	k.paramstore.Set(ctx, key, value)
}

// GetParams returns all current parameters as a types.Params instance.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetRollingSeed sets the rolling seed value to be provided value.
func (k Keeper) SetRollingSeed(ctx sdk.Context, rollingSeed []byte) {
	ctx.KVStore(k.storeKey).Set(types.RollingSeedStoreKey, rollingSeed)
}

// GetRollingSeed returns the current rolling seed value.
func (k Keeper) GetRollingSeed(ctx sdk.Context) []byte {
	return ctx.KVStore(k.storeKey).Get(types.RollingSeedStoreKey)
}

// SetRequestCount sets the number of request count to the given value. Useful for genesis state.
func (k Keeper) SetRequestCount(ctx sdk.Context, count int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.RequestCountStoreKey, k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.Int64Value{Value: count}))
}

// GetRequestCount returns the current number of all requests ever exist.
func (k Keeper) GetRequestCount(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get(types.RequestCountStoreKey)
	intV := gogotypes.Int64Value{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &intV)
	return intV.GetValue()
}

// SetRequestLastExpired sets the ID of the last expired request.
func (k Keeper) SetRequestLastExpired(ctx sdk.Context, id types.RequestID) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.RequestLastExpiredStoreKey, k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.Int64Value{Value: int64(id)}))
}

// GetRequestLastExpired returns the ID of the last expired request.
func (k Keeper) GetRequestLastExpired(ctx sdk.Context) types.RequestID {
	bz := ctx.KVStore(k.storeKey).Get(types.RequestLastExpiredStoreKey)
	intV := gogotypes.Int64Value{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &intV)
	return types.RequestID(intV.GetValue())
}

// GetNextRequestID increments and returns the current number of requests.
func (k Keeper) GetNextRequestID(ctx sdk.Context) types.RequestID {
	requestNumber := k.GetRequestCount(ctx)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.Int64Value{Value: requestNumber + 1})
	ctx.KVStore(k.storeKey).Set(types.RequestCountStoreKey, bz)
	return types.RequestID(requestNumber + 1)
}

// SetDataSourceCount sets the number of data source count to the given value.
func (k Keeper) SetDataSourceCount(ctx sdk.Context, count int64) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.Int64Value{Value: count})
	ctx.KVStore(k.storeKey).Set(types.DataSourceCountStoreKey, bz)
}

// GetDataSourceCount returns the current number of all data sources ever exist.
func (k Keeper) GetDataSourceCount(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get(types.DataSourceCountStoreKey)
	intV := gogotypes.Int64Value{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &intV)
	return intV.GetValue()
}

// GetNextDataSourceID increments and returns the current number of data sources.
func (k Keeper) GetNextDataSourceID(ctx sdk.Context) types.DataSourceID {
	dataSourceCount := k.GetDataSourceCount(ctx)
	k.SetDataSourceCount(ctx, dataSourceCount+1)
	return types.DataSourceID(dataSourceCount + 1)
}

// SetOracleScriptCount sets the number of oracle script count to the given value.
func (k Keeper) SetOracleScriptCount(ctx sdk.Context, count int64) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.Int64Value{Value: count})
	ctx.KVStore(k.storeKey).Set(types.OracleScriptCountStoreKey, bz)
}

// GetOracleScriptCount returns the current number of all oracle scripts ever exist.
func (k Keeper) GetOracleScriptCount(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get(types.OracleScriptCountStoreKey)
	intV := gogotypes.Int64Value{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &intV)
	return intV.GetValue()
}

// GetNextOracleScriptID increments and returns the current number of oracle scripts.
func (k Keeper) GetNextOracleScriptID(ctx sdk.Context) types.OracleScriptID {
	oracleScriptCount := k.GetOracleScriptCount(ctx)
	k.SetOracleScriptCount(ctx, oracleScriptCount+1)
	return types.OracleScriptID(oracleScriptCount + 1)
}

// GetFile loads the file from the file storage. Panics if the file does not exist.
func (k Keeper) GetFile(name string) []byte {
	return k.fileCache.MustGetFile(name)
}
