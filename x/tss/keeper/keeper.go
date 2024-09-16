package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	authzKeeper       types.AuthzKeeper
	rollingseedKeeper types.RollingseedKeeper

	contentRouter *types.ContentRouter
	cbRouter      *types.CallbackRouter
	authority     string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authzKeeper types.AuthzKeeper,
	rollingseedKeeper types.RollingseedKeeper,
	contentRouter *types.ContentRouter,
	cbRouter *types.CallbackRouter,
	authority string,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid tss authority address: %w", err))
	}

	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		authzKeeper:       authzKeeper,
		rollingseedKeeper: rollingseedKeeper,
		contentRouter:     contentRouter,
		cbRouter:          cbRouter,
		authority:         authority,
	}
}

// GetAuthority returns the x/tss module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// CheckIsGrantee checks if the granter granted permissions to the grantee.
func (k Keeper) CheckIsGrantee(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) bool {
	for _, msg := range types.GetGrantMsgTypes() {
		cap, _ := k.authzKeeper.GetAuthorization(
			ctx,
			grantee,
			granter,
			msg,
		)

		if cap == nil {
			return false
		}
	}

	return true
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
