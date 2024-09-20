package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/bandprotocol/chain/v3/x/bank/types"
)

// WrappedBankKeeper encapsulates the underlying bank keeper and overrides
// its BurnCoins function to send the coins to the community pool instead of
// just destroying them.
//
// Note that distrKeeper keeps the reference to the distr module keeper.
// Due to the circular dependency between bank-distr, distrKeeper
// cannot be initialized when the struct is created. Rather, SetDistrKeeper and SetAccountKeeper
// are expected to be called to set `distrKeeper` and `accountKeeper` respectively.
type WrappedBankKeeper struct {
	bankkeeper.Keeper

	distrKeeper   types.DistributionKeeper
	accountKeeper types.AccountKeeper
	logger        log.Logger
}

// NewWrappedBankKeeperBurnToCommunityPool creates a new instance of WrappedBankKeeper
// with its distrKeeper and accountKeeper members set to nil.
func NewWrappedBankKeeperBurnToCommunityPool(
	bk bankkeeper.Keeper,
	ak types.AccountKeeper,
	logger log.Logger,
) *WrappedBankKeeper {
	return &WrappedBankKeeper{
		Keeper:        bk,
		distrKeeper:   nil,
		accountKeeper: ak,
		logger:        logger.With(log.ModuleKey, "x/wrappedbank"),
	}
}

// SetDistrKeeper sets distr module keeper for this WrappedBankKeeper instance.
func (k *WrappedBankKeeper) SetDistrKeeper(dk types.DistributionKeeper) {
	k.distrKeeper = dk
}

// BurnCoins moves the specified amount of coins from the given module name to
// the community pool. The total bank of the coins will not change.
func (k WrappedBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	fmt.Println("Burn!")
	// If distrKeeper is not set OR we want to burn coins in distr itself, we will
	// just use the original BurnCoins function.

	if k.distrKeeper == nil || moduleName == distrtypes.ModuleName {
		return k.Keeper.BurnCoins(ctx, moduleName, amt)
	}

	// Create the account if it doesn't yet exist.
	acc := k.accountKeeper.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(sdkerrors.ErrUnknownAddress.Wrapf(
			"module account %s does not exist", moduleName,
		))
	}

	if !acc.HasPermission(authtypes.Burner) {
		panic(sdkerrors.ErrUnauthorized.Wrapf(
			"module account %s does not have permissions to burn tokens",
			moduleName,
		))
	}

	// Instead of burning coins, we send them to the community pool.
	if err := k.distrKeeper.FundCommunityPool(ctx, amt, acc.GetAddress()); err != nil {
		return err
	}

	k.logger.Info(fmt.Sprintf(
		"sent %s from %s module account to community pool", amt.String(), moduleName,
	))
	return nil
}
