package band

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcante "github.com/cosmos/ibc-go/v5/modules/core/ante"
	"github.com/cosmos/ibc-go/v5/modules/core/keeper"

	"github.com/bandprotocol/chain/v2/x/globalfee/feechecker"
	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	OracleKeeper      *oraclekeeper.Keeper
	IBCKeeper         *keeper.Keeper
	GlobalFeeSubspace paramtypes.Subspace
	StakingSubspace   paramtypes.Subspace
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.OracleKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "oracle keeper is required for AnteHandler")
	}
	if options.IBCKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "IBC keeper is required for AnteHandler")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	if options.TxFeeChecker == nil {
		if options.GlobalFeeSubspace.Name() == "" {
			return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "globalfee param store is required for AnteHandler")
		}
		if options.StakingSubspace.Name() == "" {
			return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "staking param store is required for AnteHandler")
		}

		feeChecker := feechecker.NewFeeChecker(
			options.OracleKeeper,
			options.GlobalFeeSubspace,
			options.StakingSubspace,
		)
		options.TxFeeChecker = feeChecker.CheckTxFeeWithMinGasPrices
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called firs(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeegrantKeeper,
			options.TxFeeChecker,
		),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
