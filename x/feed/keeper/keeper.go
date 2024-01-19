package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feed/types"
)

type Keeper struct {
	storeKey     storetypes.StoreKey
	cdc          codec.BinaryCodec
	oracleKeeper types.OracleKeeper

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	oracleKeeper types.OracleKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		oracleKeeper: oracleKeeper,
		authority:    authority,
	}
}

// GetAuthority returns the x/feed module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
