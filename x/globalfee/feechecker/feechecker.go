package feechecker

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/bandprotocol/chain/v3/x/globalfee/keeper"
)

type FeeChecker struct {
	GlobalfeeKeeper *keeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
}

func NewFeeChecker(
	globalfeeKeeper *keeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) FeeChecker {
	return FeeChecker{
		GlobalfeeKeeper: globalfeeKeeper,
		StakingKeeper:   stakingKeeper,
	}
}

// CheckTxFee is responsible for verifying whether a transaction contains the necessary fee.
func (fc FeeChecker) CheckTxFee(
	ctx sdk.Context,
	tx sdk.Tx,
) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	bondDenom, err := fc.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, 0, err
	}
	priority := getTxPriority(feeCoins, int64(gas), bondDenom)

	// Ensure that the provided fees meet minimum-gas-prices and globalFees,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if !ctx.IsCheckTx() {
		return feeCoins, priority, nil
	}

	minGasPrices := getMinGasPrices(ctx)
	globalMinGasPrices, err := fc.GetGlobalMinGasPrices(ctx)
	if err != nil {
		return nil, 0, err
	}

	allGasPrices := CombinedGasPricesRequirement(minGasPrices, globalMinGasPrices)

	// Calculate all fees from all gas prices
	var allFees sdk.Coins
	if !allGasPrices.IsZero() {
		glDec := sdkmath.LegacyNewDec(int64(gas))
		for _, gp := range allGasPrices {
			if !gp.IsZero() {
				fee := gp.Amount.Mul(glDec)
				allFees = append(allFees, sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt()))
			}
		}
	}

	if !allFees.IsZero() && !feeCoins.IsAnyGTE(allFees) {
		return nil, 0, sdkerrors.ErrInsufficientFee.Wrapf(
			"insufficient fees; got: %s required: %s",
			feeCoins,
			allFees,
		)
	}

	return feeCoins, priority, nil
}

// GetGlobalMinGasPrices returns global min gas prices
func (fc FeeChecker) GetGlobalMinGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	globalMinGasPrices := fc.GlobalfeeKeeper.GetParams(ctx).MinimumGasPrices
	if len(globalMinGasPrices) != 0 {
		return globalMinGasPrices.Sort(), nil
	}
	// global fee is empty set, set global fee to 0uband (bondDenom)
	globalMinGasPrices, err := fc.DefaultZeroGlobalFee(ctx)
	if err != nil {
		return globalMinGasPrices, err
	}

	return globalMinGasPrices.Sort(), nil
}

// DefaultZeroGlobalFee returns a zero coin with the staking module bond denom
func (fc FeeChecker) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom, err := fc.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdkmath.LegacyNewDec(0))}, nil
}
