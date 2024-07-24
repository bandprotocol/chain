package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// Keeper of the x/tunnel store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	feedsKeeper   types.FeedsKeeper
	bandtssKeeper types.BandtssKeeper
	channelKeeper types.ChannelKeeper
	scopedKeeper  types.ScopedKeeper

	authority string
}

// NewKeeper creates a new tunnel Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	feedsKeeper types.FeedsKeeper,
	bandtssKeeper types.BandtssKeeper,
	channelKeeper types.ChannelKeeper,
	scopedKeeper types.ScopedKeeper,
	authority string,
) Keeper {
	// ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      key,
		authKeeper:    authKeeper,
		bankKeeper:    bankKeeper,
		feedsKeeper:   feedsKeeper,
		bandtssKeeper: bandtssKeeper,
		channelKeeper: channelKeeper,
		scopedKeeper:  scopedKeeper,
		authority:     authority,
	}
}

// GetTunnelAccount returns the tunnel ModuleAccount
func (k Keeper) GetTunnelAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetModuleBalance returns the balance of the tunnel ModuleAccount
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, k.GetTunnelAccount(ctx).GetAddress())
}

// SetModuleAccount sets a module account in the account keeper.
func (k Keeper) SetModuleAccount(ctx sdk.Context, acc authtypes.ModuleAccountI) {
	k.authKeeper.SetModuleAccount(ctx, acc)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
