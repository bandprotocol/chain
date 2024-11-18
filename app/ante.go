package band

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

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
	Cdc                   codec.Codec
	AuthzKeeper           *authzkeeper.Keeper
	OracleKeeper          *oraclekeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	GlobalfeeKeeper       *globalfeekeeper.Keeper
	TSSKeeper             *tsskeeper.Keeper
	BandtssKeeper         *bandtsskeeper.Keeper
	FeedsKeeper           *feedskeeper.Keeper
	TXCounterStoreService corestoretypes.KVStoreService
	WasmConfig            *wasmtypes.WasmConfig
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
		wasmkeeper.NewLimitSimulationGasDecorator(
			options.WasmConfig.SimulationGasLimit,
		), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
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
