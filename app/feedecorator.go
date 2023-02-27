package band

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MinFeeDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MinFeeDecorator.
type MinFeeDecorator struct {
}

func NewMinFeeDecorator() MinFeeDecorator {
	return MinFeeDecorator{}
}

func (mfd MinFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	if len(feeCoins) > 1 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "too many fee coins. only accepts fees in one denom")
	}

	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
		// Determine if these fees are sufficient for the tx to pass.
		// Once ABCI++ Process Proposal lands, we can have block validity conditions enforce this.
		minBaseGasPrice := mfd.getMinBaseGasPrice(ctx, feeTx)

		// If minBaseGasPrice is zero, then we don't need to check the fee. Continue
		if minBaseGasPrice.IsZero() {
			return next(ctx, tx, simulate)
		}
		// You should only be able to pay with one fee token in a single tx
		if len(feeCoins) != 1 {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee,
				"Expected 1 fee denom attached, got %d", len(feeCoins))
		}
		// The minimum base gas price is in uband, convert the fee denom's worth to uband terms.
		// Then compare if its sufficient for paying the tx fee.
		err = mfd.isSufficientFee(minBaseGasPrice, gas, feeCoins[0])
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (mfd MinFeeDecorator) getMinBaseGasPrice(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Dec {
	// In block execution (DeliverTx), its set to the governance decided upon consensus min fee.
	minBaseGasPrice := ConsensusMinFee

	// If we are in genesis, then we actually override all of the above, to set it to 0.
	if ctx.BlockHeight() == 0 {
		minBaseGasPrice = sdk.ZeroDec()
	}
	return minBaseGasPrice
}

func (mfd MinFeeDecorator) isSufficientFee(minBaseGasPrice sdk.Dec, gasRequested uint64, feeCoin sdk.Coin) error {
	// Determine the required fees by multiplying the required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(gasRequested))
	requiredFee := sdk.NewCoin(Denom, minBaseGasPrice.Mul(glDec).Ceil().RoundInt())

	// check to ensure that the feeCoin should always be greater than or equal to the requireBaseFee
	if !(feeCoin.IsGTE(requiredFee)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s. required: %s", feeCoin, requiredFee)
	}

	return nil
}
