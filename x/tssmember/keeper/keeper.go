package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/bandprotocol/chain/v2/x/tssmember/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	tssKeeper     types.TSSKeeper
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	distrKeeper   types.DistrKeeper

	authority        string
	feeCollectorName string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistrKeeper,
	tssKeeper types.TSSKeeper,
	authority string,
	feeCollectorName string,
) Keeper {
	// ensure TSS module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		paramSpace:       paramSpace,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		stakingKeeper:    stakingKeeper,
		distrKeeper:      distrKeeper,
		tssKeeper:        tssKeeper,
		authority:        authority,
		feeCollectorName: feeCollectorName,
	}
}

// GetTSSAccount returns the TSS ModuleAccount
func (k Keeper) GetTSSMemberAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
