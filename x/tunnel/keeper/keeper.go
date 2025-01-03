package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	authKeeper     types.AccountKeeper
	bankKeeper     types.BankKeeper
	feedsKeeper    types.FeedsKeeper
	bandtssKeeper  types.BandtssKeeper
	channelKeeper  types.ChannelKeeper
	ics4Wrapper    types.ICS4Wrapper
	portKeeper     types.PortKeeper
	scopedKeeper   types.ScopedKeeper
	transferKeeper types.TransferKeeper

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
	ics4Wrapper types.ICS4Wrapper,
	portKeeper types.PortKeeper,
	scopedKeeper types.ScopedKeeper,
	transferKeeper types.TransferKeeper,
	authority string,
) Keeper {
	// ensure tunnel module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		authKeeper:     authKeeper,
		bankKeeper:     bankKeeper,
		feedsKeeper:    feedsKeeper,
		bandtssKeeper:  bandtssKeeper,
		channelKeeper:  channelKeeper,
		ics4Wrapper:    ics4Wrapper,
		portKeeper:     portKeeper,
		scopedKeeper:   scopedKeeper,
		transferKeeper: transferKeeper,
		authority:      authority,
	}
}

// GetAuthority returns the x/tunnel module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetTunnelAccount returns the tunnel ModuleAccount
func (k Keeper) GetTunnelAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetModuleBalance returns the balance of the tunnel ModuleAccount
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, k.GetTunnelAccount(ctx).GetAddress())
}

// SetModuleAccount sets a module account in the account keeper.
func (k Keeper) SetModuleAccount(ctx sdk.Context, acc sdk.ModuleAccountI) {
	k.authKeeper.SetModuleAccount(ctx, acc)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
