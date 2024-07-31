package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

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

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the oracle module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
