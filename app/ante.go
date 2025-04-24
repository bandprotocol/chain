package band

import (
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/bandprotocol/chain/v3/app/mempool"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	feedskeeper "github.com/bandprotocol/chain/v3/x/feeds/keeper"
	"github.com/bandprotocol/chain/v3/x/globalfee/feechecker"
	globalfeekeeper "github.com/bandprotocol/chain/v3/x/globalfee/keeper"
	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	Cdc                     codec.Codec
	AuthzKeeper             *authzkeeper.Keeper
	OracleKeeper            *oraclekeeper.Keeper
	IBCKeeper               *ibckeeper.Keeper
	StakingKeeper           *stakingkeeper.Keeper
	GlobalfeeKeeper         *globalfeekeeper.Keeper
	TSSKeeper               *tsskeeper.Keeper
	BandtssKeeper           *bandtsskeeper.Keeper
	FeedsKeeper             *feedskeeper.Keeper
	IgnoreDecoratorMatchFns []mempool.TxMatchFn
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.Cdc == nil {
		return nil, sdkerrors.ErrLogic.Wrap("codec is required for AnteHandler")
	}
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign mode handler is required for ante builder")
	}
	if options.AuthzKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("authz keeper is required for AnteHandler")
	}
	if options.OracleKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("oracle keeper is required for AnteHandler")
	}
	if options.TSSKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("tss keeper is required for AnteHandler")
	}
	if options.BandtssKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bandtss keeper is required for AnteHandler")
	}
	if options.FeedsKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("feeds keeper is required for AnteHandler")
	}
	if options.IBCKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("IBC keeper is required for AnteHandler")
	}
	if options.StakingKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("Staking keeper is required for AnteHandler")
	}
	if options.GlobalfeeKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("Globalfee keeper is required for AnteHandler")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	if options.TxFeeChecker == nil {
		feeChecker := feechecker.NewFeeChecker(
			options.Cdc,
			options.AuthzKeeper,
			options.OracleKeeper,
			options.GlobalfeeKeeper,
			options.StakingKeeper,
			options.TSSKeeper,
			options.BandtssKeeper,
			options.FeedsKeeper,
		)
		options.TxFeeChecker = feeChecker.CheckTxFee
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewIgnoreDecorator(
			ante.NewDeductFeeDecorator(
				options.AccountKeeper,
				options.BankKeeper,
				options.FeegrantKeeper,
				options.TxFeeChecker,
			),
			options.IgnoreDecoratorMatchFns...,
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

// IgnoreDecorator is an AnteDecorator that wraps an existing AnteDecorator. It allows
// for the AnteDecorator to be ignored for specified lanes.
type IgnoreDecorator struct {
	decorator sdk.AnteDecorator
	matchFns  []mempool.TxMatchFn
}

// NewIgnoreDecorator returns a new IgnoreDecorator instance.
func NewIgnoreDecorator(decorator sdk.AnteDecorator, matchFns ...mempool.TxMatchFn) *IgnoreDecorator {
	return &IgnoreDecorator{
		decorator: decorator,
		matchFns:  matchFns,
	}
}

// NewIgnoreDecorator is a wrapper that implements the sdk.AnteDecorator interface,
// providing two execution paths for processing transactions:
//   - If a transaction matches one of the designated bypass lanes, it is forwarded
//     directly to the next AnteHandler.
//   - Otherwise, the transaction is processed using the embedded decorator’s AnteHandler.
func (ig IgnoreDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	// IgnoreDecorator is only used for check tx.
	if !ctx.IsCheckTx() {
		return ig.decorator.AnteHandle(ctx, tx, simulate, next)
	}

	cacheCtx, _ := ctx.CacheContext()
	for _, matchFn := range ig.matchFns {
		if matchFn(cacheCtx, tx) {
			return next(ctx, tx, simulate)
		}
	}

	return ig.decorator.AnteHandle(ctx, tx, simulate, next)
}
