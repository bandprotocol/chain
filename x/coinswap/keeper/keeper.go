package keeper

import (
	"fmt"
	"github.com/GeoDB-Limited/odincore/chain/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	paramSpace   params.Subspace
	supplyKeeper types.SupplyKeeper
	distrKeeper  types.DistrKeeper
	oracleKeeper types.OracleKeeper
}

func NewKeeper(
	cdc *codec.Codec,
	key sdk.StoreKey,
	subspace params.Subspace,
	sk types.SupplyKeeper,
	dk types.DistrKeeper,
	ok types.OracleKeeper) Keeper {

	if !subspace.HasKeyTable() {
		subspace = subspace.WithKeyTable(types.ParamKeyTable())
	}
	return Keeper{
		cdc:          cdc,
		storeKey:     key,
		paramSpace:   subspace,
		supplyKeeper: sk,
		distrKeeper:  dk,
		oracleKeeper: ok,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetDecParam returns the parameter as specified by key as sdk.Dec
func (k Keeper) GetDecParam(ctx sdk.Context, key []byte) (res sdk.Dec) {
	k.paramSpace.Get(ctx, key, &res)
	return res
}

// GetDecParam returns the parameter as specified by key as types.ValidExchanges
func (k Keeper) GetValidExchangesParam(ctx sdk.Context, key []byte) (res types.ValidExchanges) {
	k.paramSpace.Get(ctx, key, &res)
	return res
}

// SetParam saves the given key-value parameter to the store.
func (k Keeper) SetParams(ctx sdk.Context, value types.Params) {
	k.paramSpace.SetParamSet(ctx, &value)
}

// GetParams returns all current parameters as a types.Params instance.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetInitialRate(ctx sdk.Context, value sdk.Dec) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(value)
	ctx.KVStore(k.storeKey).Set(types.InitialRateStoreKey, bz)
}

func (k Keeper) GetInitialRate(ctx sdk.Context) (rate sdk.Dec) {
	bz := ctx.KVStore(k.storeKey).Get(types.InitialRateStoreKey)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rate)
	return rate
}

func (k Keeper) GetRateMultiplier(ctx sdk.Context) (multiplier sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyRateMultiplier, &multiplier)
	return multiplier
}
