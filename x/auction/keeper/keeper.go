package keeper

import (
	"fmt"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey       sdk.StoreKey
	cdc            codec.BinaryMarshaler
	paramstore     paramstypes.Subspace
	bankKeeper     auctiontypes.BankKeeper
	distrKeeper    auctiontypes.DistrKeeper
	oracleKeeper   auctiontypes.OracleKeeper
	coinswapKeeper auctiontypes.CoinswapKeeper
}

func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	subspace paramstypes.Subspace,
	bk auctiontypes.BankKeeper,
	dk auctiontypes.DistrKeeper,
	ok auctiontypes.OracleKeeper,
	ck auctiontypes.CoinswapKeeper,
) Keeper {

	if !subspace.HasKeyTable() {
		subspace = subspace.WithKeyTable(auctiontypes.ParamKeyTable())
	}
	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		paramstore:     subspace,
		bankKeeper:     bk,
		distrKeeper:    dk,
		oracleKeeper:   ok,
		coinswapKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", auctiontypes.ModuleName))
}

// SetParams saves the given key-value parameter to the store.
func (k Keeper) SetParams(ctx sdk.Context, value auctiontypes.Params) {
	k.paramstore.SetParamSet(ctx, &value)
}

// GetParams returns all current parameters as a types.Params instance.
func (k Keeper) GetParams(ctx sdk.Context) (params auctiontypes.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}
