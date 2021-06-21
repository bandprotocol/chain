package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	owasm "github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v2/pkg/filecache"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
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
	owasmVM          *owasm.Vm

	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistrKeeper
	stakingKeeper types.StakingKeeper
	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
}

// NewKeeper creates a new oracle Keeper instance.
func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	ps paramtypes.Subspace,
	fileDir string,
	feeCollectorName string,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistrKeeper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopeKeeper capabilitykeeper.ScopedKeeper,
	owasmVM *owasm.Vm,
) Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}
	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		fileCache:        filecache.New(fileDir),
		feeCollectorName: feeCollectorName,
		paramstore:       ps,
		owasmVM:          owasmVM,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		distrKeeper:      distrKeeper,
		stakingKeeper:    stakingKeeper,
		channelKeeper:    channelKeeper,
		portKeeper:       portKeeper,
		scopedKeeper:     scopeKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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

// IsBound checks if the transfer module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the ort Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort returns the portID for the transfer module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the transfer module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
