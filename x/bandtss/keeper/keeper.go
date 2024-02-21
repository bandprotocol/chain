package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistrKeeper
	stakingKeeper types.StakingKeeper
	tssKeeper     types.TSSKeeper

	authority        string
	feeCollectorName string

	router *types.Router
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistrKeeper,
	stakingKeeper types.StakingKeeper,
	tssKeeper types.TSSKeeper,
	authority string,
	feeCollectorName string,
	router *types.Router,
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
		distrKeeper:      distrKeeper,
		stakingKeeper:    stakingKeeper,
		tssKeeper:        tssKeeper,
		authority:        authority,
		feeCollectorName: feeCollectorName,
		router:           router,
	}
}

// GetTSSAccount returns the TSS ModuleAccount
func (k Keeper) GetTSSMemberAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
