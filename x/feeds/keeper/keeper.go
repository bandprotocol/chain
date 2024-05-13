package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	oracleKeeper  types.OracleKeeper
	stakingKeeper types.StakingKeeper

	authority string
}

// NewKeeper creates a new feeds Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	oracleKeeper types.OracleKeeper,
	stakingKeeper types.StakingKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		oracleKeeper:  oracleKeeper,
		stakingKeeper: stakingKeeper,
		authority:     authority,
	}
}

// GetAuthority returns the x/feeds module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// IsBondedValidator checks is the validator is in the bonded validators.
func (k Keeper) IsBondedValidator(ctx sdk.Context, valAddr string) bool {
	addr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return false
	}

	val, found := k.stakingKeeper.GetValidator(ctx, addr)
	if !found {
		return false
	}

	return val.IsBonded()
}
