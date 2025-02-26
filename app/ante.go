package band

import (
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"

	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/bandprotocol/chain/v3/app/mempool"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	feedskeeper "github.com/bandprotocol/chain/v3/x/feeds/keeper"
	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
)

// UseFeeMarketDecorator to make the integration testing easier: we can switch off its ante and post decorators with this flag
var UseFeeMarketDecorator = true

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	Cdc                    codec.Codec
	ExtensionOptionChecker ante.ExtensionOptionChecker
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	TxFeeChecker           ante.TxFeeChecker

	AccountKeeper   feemarketante.AccountKeeper
	BankKeeper      feemarketante.BankKeeper
	FeegrantKeeper  ante.FeegrantKeeper
	AuthzKeeper     *authzkeeper.Keeper
	OracleKeeper    *oraclekeeper.Keeper
	IBCKeeper       *ibckeeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
	FeeMarketKeeper *feemarketkeeper.Keeper
	TSSKeeper       *tsskeeper.Keeper
	BandtssKeeper   *bandtsskeeper.Keeper
	FeedsKeeper     *feedskeeper.Keeper

	Lanes []*mempool.Lane
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
	if options.FeeMarketKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("FeeMarket keeper is required for AnteHandler")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	if UseFeeMarketDecorator {
		anteDecorators = append(anteDecorators,
			NewIgnoreDecorator(
				feemarketante.NewFeeMarketCheckDecorator(
					options.AccountKeeper,
					options.BankKeeper,
					options.FeegrantKeeper,
					options.FeeMarketKeeper,
					ante.NewDeductFeeDecorator(
						options.AccountKeeper,
						options.BankKeeper,
						options.FeegrantKeeper,
						options.TxFeeChecker)),
				options.Lanes...,
			),
		)
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// IgnoreDecorator is an AnteDecorator that wraps an existing AnteDecorator. It allows
// for the AnteDecorator to be ignored for specified lanes.
type IgnoreDecorator struct {
	decorator sdk.AnteDecorator
	lanes     []*mempool.Lane
}

// NewIgnoreDecorator returns a new IgnoreDecorator instance.
func NewIgnoreDecorator(decorator sdk.AnteDecorator, lanes ...*mempool.Lane) *IgnoreDecorator {
	return &IgnoreDecorator{
		decorator: decorator,
		lanes:     lanes,
	}
}

// AnteHandle implements the sdk.AnteDecorator interface. If the transaction belongs to
// one of the lanes, the next AnteHandler is called. Otherwise, the decorator's AnteHandler
// is called.
func (sd IgnoreDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	cacheCtx, _ := ctx.CacheContext()
	for _, lane := range sd.lanes {
		if lane.Match(cacheCtx, tx) {
			return next(ctx, tx, simulate)
		}
	}

	return sd.decorator.AnteHandle(ctx, tx, simulate, next)
}
