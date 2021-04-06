package keeper

import (
	"fmt"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          codec.BinaryMarshaler
	paramstore   paramstypes.Subspace
	bankKeeper   coinswaptypes.BankKeeper
	distrKeeper  coinswaptypes.DistrKeeper
	oracleKeeper coinswaptypes.OracleKeeper
}

func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	subspace paramstypes.Subspace,
	ak coinswaptypes.AccountKeeper,
	bk coinswaptypes.BankKeeper,
	dk coinswaptypes.DistrKeeper,
	ok coinswaptypes.OracleKeeper) Keeper {

	if !subspace.HasKeyTable() {
		subspace = subspace.WithKeyTable(coinswaptypes.ParamKeyTable())
	}
	return Keeper{
		cdc:          cdc,
		storeKey:     key,
		paramstore:   subspace,
		bankKeeper:   bk,
		distrKeeper:  dk,
		oracleKeeper: ok,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", coinswaptypes.ModuleName))
}

// GetDecParam returns the parameter as specified by key as sdk.Dec
func (k Keeper) GetDecParam(ctx sdk.Context, key []byte) (res sdk.Dec) {
	k.paramstore.Get(ctx, key, &res)
	return res
}

// GetDecParam returns the parameter as specified by key as types.ValidExchanges
func (k Keeper) GetValidExchangesParam(ctx sdk.Context, key []byte) (res coinswaptypes.ValidExchanges) {
	k.paramstore.Get(ctx, key, &res)
	return res
}

// SetParam saves the given key-value parameter to the store.
func (k Keeper) SetParams(ctx sdk.Context, value coinswaptypes.Params) {
	k.paramstore.SetParamSet(ctx, &value)
}

// GetParams returns all current parameters as a types.Params instance.
func (k Keeper) GetParams(ctx sdk.Context) (params coinswaptypes.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetInitialRate(ctx sdk.Context, value sdk.Dec) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.StringValue{Value: value.String()})
	ctx.KVStore(k.storeKey).Set(coinswaptypes.InitialRateStoreKey, bz)
}

func (k Keeper) GetInitialRate(ctx sdk.Context) (rate sdk.Dec) {
	bz := ctx.KVStore(k.storeKey).Get(coinswaptypes.InitialRateStoreKey)
	var rawRate gogotypes.StringValue
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rawRate)
	rate = sdk.MustNewDecFromStr(rawRate.Value)
	return rate
}

func (k Keeper) GetRateMultiplier(ctx sdk.Context) (multiplier sdk.Dec) {
	k.paramstore.Get(ctx, coinswaptypes.KeyRateMultiplier, &multiplier)
	return multiplier
}
