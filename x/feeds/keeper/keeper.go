package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	oracleKeeper  types.OracleKeeper
	stakingKeeper types.StakingKeeper
	restakeKeeper types.RestakeKeeper
	authzKeeper   types.AuthzKeeper

	authority string
}

// NewKeeper creates a new feeds Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	oracleKeeper types.OracleKeeper,
	stakingKeeper types.StakingKeeper,
	restakeKeeper types.RestakeKeeper,
	authzKeeper types.AuthzKeeper,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		oracleKeeper:  oracleKeeper,
		stakingKeeper: stakingKeeper,
		restakeKeeper: restakeKeeper,
		authzKeeper:   authzKeeper,
		authority:     authority,
	}
}

// GetAuthority returns the x/feeds module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// IsBondedValidator checks is the validator is in the bonded validators.
func (k Keeper) IsBondedValidator(ctx sdk.Context, addr sdk.ValAddress) bool {
	val, err := k.stakingKeeper.GetValidator(ctx, addr)
	if err != nil {
		return false
	}

	return val.IsBonded()
}

// IsFeeder checks if the given address has been granted as a feeder by the given validator
func (k Keeper) IsFeeder(ctx sdk.Context, validator sdk.ValAddress, feeder sdk.AccAddress) bool {
	cap, _ := k.authzKeeper.GetAuthorization(
		ctx,
		feeder,
		sdk.AccAddress(validator),
		sdk.MsgTypeURL(&types.MsgSubmitSignalPrices{}),
	)
	return cap != nil
}
