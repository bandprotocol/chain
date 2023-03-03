package band

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// DeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	accountKeeper  ante.AccountKeeper
	bankKeeper     authtypes.BankKeeper
	feegrantKeeper ante.FeegrantKeeper
	oracleKeeper   oraclekeeper.Keeper
	txFeeChecker   ante.TxFeeChecker
}

func NewDeductFeeDecorator(ak ante.AccountKeeper, bk authtypes.BankKeeper, fk ante.FeegrantKeeper, ork oraclekeeper.Keeper, tfc ante.TxFeeChecker) DeductFeeDecorator {
	if tfc == nil {
		tfc = CheckTxFeeWithValidatorMinGasPrices
	}

	return DeductFeeDecorator{
		accountKeeper:  ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		oracleKeeper:   ork,
		txFeeChecker:   tfc,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	// check is report tx
	isReportTx, err := dfd.checkReportTx(ctx, tx)
	if err != nil {
		return ctx, err
	}

	var priority int64

	fee := feeTx.GetFee()
	gas := feeTx.GetGas()

	// check min gas price
	if !isReportTx {
		if err := checkMinGasPrice(ctx, feeTx, fee, gas); err != nil {
			return ctx, err
		}
	}

	minGas := ctx.MinGasPrices()
	if !simulate {
		if isReportTx {
			fee, _, err = dfd.txFeeChecker(ctx.WithMinGasPrices(sdk.DecCoins{}), tx)

			// overwrite to high priority if msg is a reportMsg
			priority = int64(math.MaxInt64)
		} else {
			fee, priority, err = dfd.txFeeChecker(ctx, tx)
		}
		if err != nil {
			return ctx, err
		}
	}
	if err := dfd.checkDeductFee(ctx, tx, fee); err != nil {
		return ctx, err
	}

	newCtx, err := next(ctx.WithPriority(priority), tx, simulate)
	return newCtx.WithMinGasPrices(minGas), err
}

func (dfd DeductFeeDecorator) checkDeductFee(ctx sdk.Context, sdkTx sdk.Tx, fee sdk.Coins) error {
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", authtypes.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, sdkTx.GetMsgs())
			if err != nil {
				return sdkerrors.Wrapf(err, "%s does not not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !fee.IsZero() {
		err := DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, fee)
		if err != nil {
			return err
		}
	}

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

func (dfd DeductFeeDecorator) checkReportTx(ctx sdk.Context, tx sdk.Tx) (bool, error) {
	isValidReportTx := true

	for _, msg := range tx.GetMsgs() {
		// Check direct report msg
		if dr, ok := msg.(*types.MsgReportData); ok {
			// Check if it's not valid report msg, discard this transaction
			if err := checkValidReportMsg(ctx, dfd.oracleKeeper, dr); err != nil {
				return false, err
			}
		} else {
			isValid, err := checkExecMsgReportFromReporter(ctx, dfd.oracleKeeper, msg)
			if err != nil {
				return false, err
			}

			isValidReportTx = isValid
		}
	}
	if isValidReportTx {
		return true, nil
	}
	return false, nil
}

func checkValidReportMsg(ctx sdk.Context, oracleKeeper oraclekeeper.Keeper, r *types.MsgReportData) error {
	validator, err := sdk.ValAddressFromBech32(r.Validator)
	if err != nil {
		return err
	}
	report := types.NewReport(validator, false, r.RawReports)
	return oracleKeeper.CheckValidReport(ctx, r.RequestID, report)
}

func checkExecMsgReportFromReporter(ctx sdk.Context, oracleKeeper oraclekeeper.Keeper, msg sdk.Msg) (bool, error) {
	// Check is the MsgExec from reporter
	me, ok := msg.(*authz.MsgExec)
	if !ok {
		return false, nil
	}

	// If cannot get message, then pretend as non-free transaction
	msgs, err := me.GetMessages()
	if err != nil {
		return false, err
	}

	grantee, err := sdk.AccAddressFromBech32(me.Grantee)
	if err != nil {
		return false, err
	}

	allValidReportMsg := true
	for _, m := range msgs {
		r, ok := m.(*types.MsgReportData)
		// If this is not report msg, skip other msgs on this exec msg
		if !ok {
			allValidReportMsg = false
			break
		}

		// Fail to parse validator, then discard this transaction
		validator, err := sdk.ValAddressFromBech32(r.Validator)
		if err != nil {
			return false, err
		}

		// If this grantee is not a reporter of validator, then discard this transaction
		if !oracleKeeper.IsReporter(ctx, validator, grantee) {
			return false, sdkerrors.ErrUnauthorized.Wrap("authorization not found")
		}

		// Check if it's not valid report msg, discard this transaction
		if err := checkValidReportMsg(ctx, oracleKeeper, r); err != nil {
			return false, err
		}
	}
	// If this exec msg has other non-report msg, disable feeless and skip other msgs in tx
	if !allValidReportMsg {
		return false, nil
	}

	return true, nil
}

func checkMinGasPrice(ctx sdk.Context, feeTx sdk.FeeTx, feeCoins sdk.Coins, gas uint64) error {
	// Determine if these fees are sufficient for the tx to pass.
	// Once ABCI++ Process Proposal lands, we can have block validity conditions enforce this.
	minBaseGasPrice := getMinBaseGasPrice(ctx, feeTx)

	// If minBaseGasPrice is zero, then we don't need to check the fee. Continue
	if minBaseGasPrice.IsZero() {
		return nil
	}
	// You should only be able to pay with one fee token in a single tx
	if len(feeCoins) != 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee,
			"Expected 1 fee denom attached, got %d", len(feeCoins))
	}
	// The minimum base gas price is in uband, convert the fee denom's worth to uband terms.
	// Then compare if its sufficient for paying the tx fee.
	err := isSufficientFee(minBaseGasPrice, gas, feeCoins[0])
	if err != nil {
		return err
	}

	return nil
}

func getMinBaseGasPrice(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Dec {
	// In block execution (DeliverTx), its set to the governance decided upon consensus min fee.
	minBaseGasPrice := ConsensusMinFee

	// If we are in genesis, then we actually override all of the above, to set it to 0.
	if ctx.BlockHeight() == 0 {
		minBaseGasPrice = sdk.ZeroDec()
	}
	return minBaseGasPrice
}

func isSufficientFee(minBaseGasPrice sdk.Dec, gas uint64, feeCoin sdk.Coin) error {
	// Determine the required fees by multiplying the required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(gas))
	fee := minBaseGasPrice.Mul(glDec)
	requiredFee := sdk.NewCoin(Denom, fee.Ceil().RoundInt())
	// check to ensure that the feeCoin should always be greater than or equal to the requireBaseFee
	if !(feeCoin.IsGTE(requiredFee)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s. required: %s", feeCoin, requiredFee)
	}

	return nil
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper authtypes.BankKeeper, ctx sdk.Context, acc authtypes.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authtypes.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
