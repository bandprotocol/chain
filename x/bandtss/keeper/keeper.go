package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	authzKeeper   types.AuthzKeeper
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistrKeeper
	stakingKeeper types.StakingKeeper
	tssKeeper     types.TSSKeeper

	authority        string
	feeCollectorName string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authzKeeper types.AuthzKeeper,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistrKeeper,
	stakingKeeper types.StakingKeeper,
	tssKeeper types.TSSKeeper,
	authority string,
	feeCollectorName string,
) Keeper {
	// ensure bandtss module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid bandtss authority address: %w", err))
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		authzKeeper:      authzKeeper,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		distrKeeper:      distrKeeper,
		stakingKeeper:    stakingKeeper,
		tssKeeper:        tssKeeper,
		authority:        authority,
		feeCollectorName: feeCollectorName,
	}
}

// GetBandtssAccount returns the bandtss ModuleAccount
func (k Keeper) GetBandtssAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetModuleBalance returns the balance of the bandtss ModuleAccount
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, k.GetBandtssAccount(ctx).GetAddress())
}

// SetModuleAccount sets a module account in the account keeper.
func (k Keeper) SetModuleAccount(ctx sdk.Context, acc sdk.ModuleAccountI) {
	k.authKeeper.SetModuleAccount(ctx, acc)
}

// Logger gets logger object.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetCurrentGroup sets a current group information of the bandtss module.
func (k Keeper) SetCurrentGroup(ctx sdk.Context, currentGroup types.CurrentGroup) {
	ctx.KVStore(k.storeKey).Set(types.CurrentGroupStoreKey, k.cdc.MustMarshal(&currentGroup))
}

// GetCurrentGroup retrieves a current group information of the bandtss module.
func (k Keeper) GetCurrentGroup(ctx sdk.Context) types.CurrentGroup {
	bz := ctx.KVStore(k.storeKey).Get(types.CurrentGroupStoreKey)
	if bz == nil {
		return types.CurrentGroup{}
	}

	var currentGroup types.CurrentGroup
	k.cdc.MustUnmarshal(bz, &currentGroup)
	return currentGroup
}

// CheckIsGrantee checks if the granter granted permissions to the grantee.
func (k Keeper) CheckIsGrantee(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) bool {
	for _, msg := range types.GetGrantMsgTypes() {
		cap, _ := k.authzKeeper.GetAuthorization(ctx, grantee, granter, msg)
		if cap == nil {
			return false
		}
	}

	return true
}
