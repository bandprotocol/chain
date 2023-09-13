package band

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/bandprotocol/chain/v2/x/globalfee/feechecker"
	globalfeekeeper "github.com/bandprotocol/chain/v2/x/globalfee/keeper"
	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	OracleKeeper      *oraclekeeper.Keeper
	IBCKeeper         *ibckeeper.Keeper
	GlobalFeeSubspace paramtypes.Subspace
	StakingKeeper     *stakingkeeper.Keeper
	GlobalfeeKeeper   *globalfeekeeper.Keeper
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
	if options.GlobalfeeKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "Globalfee keeper is required for AnteHandler")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	if options.TxFeeChecker == nil {
		if options.GlobalFeeSubspace.Name() == "" {
			return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "globalfee param store is required for AnteHandler")
		}

		feeChecker := feechecker.NewFeeChecker(
			options.OracleKeeper,
			options.GlobalfeeKeeper,
			options.StakingKeeper,
		)
		options.TxFeeChecker = feeChecker.CheckTxFeeWithMinGasPrices
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first(),
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
