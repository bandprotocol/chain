package keeper

import (
	"fmt"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey       sdk.StoreKey
	cdc            codec.BinaryMarshaler
	paramstore     paramstypes.Subspace
	authKeeper     auctiontypes.AccountKeeper
	coinswapKeeper auctiontypes.CoinswapKeeper
}

func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	subspace paramstypes.Subspace,
	ak auctiontypes.AccountKeeper,
	ck auctiontypes.CoinswapKeeper,
) Keeper {
	// ensure auction module account is set
	if addr := ak.GetModuleAddress(auctiontypes.ModuleName); addr == nil {
		panic("the auction module account has not been set")
	}

	if !subspace.HasKeyTable() {
		subspace = subspace.WithKeyTable(auctiontypes.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		paramstore:     subspace,
		authKeeper:     ak,
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

// GetThreshold returns auction threshold parameter
func (k Keeper) GetThreshold(ctx sdk.Context) sdk.Coins {
	params := k.GetParams(ctx)
	return params.Threshold
}

// GetAuctionAccount returns the auction ModuleAccount
func (k Keeper) GetAuctionAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, auctiontypes.ModuleName)
}

// SetAuctionAccount sets the module account
func (k Keeper) SetAuctionAccount(ctx sdk.Context, moduleAcc authtypes.ModuleAccountI) {
	k.authKeeper.SetModuleAccount(ctx, moduleAcc)
}

// BuyCoins buys geo for odin from data providers pool
func (k Keeper) BuyCoins(ctx sdk.Context) error {
	moduleAcc := k.GetAuctionAccount(ctx)
	params := k.GetParams(ctx)
	exchangeAmt := sdk.NewCoin(params.ExchangeRate.From, params.Threshold.AmountOf(params.ExchangeRate.From))
	err := k.coinswapKeeper.Exchange(ctx, exchangeAmt, params.ExchangeRate, moduleAcc.GetAddress())
	return sdkerrors.Wrap(err, "failed to exchange coins")
}
