package bandtss

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	bandtssclient "github.com/bandprotocol/chain/v3/x/bandtss/client"
	"github.com/bandprotocol/chain/v3/x/bandtss/client/cli"
	"github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// ConsensusVersion defines the current x/feeds module consensus version.
const ConsensusVersion uint64 = 1

var (
	_ module.AppModuleBasic = AppModuleBasic{}

	_ module.HasGenesis         = AppModule{}
	_ module.HasServices        = AppModule{}
	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasEndBlocker   = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic defines the basic application module used by the bandtss module.
type AppModuleBasic struct {
	cdc                    codec.BinaryCodec
	signatureOrderHandlers []bandtssclient.RequestSignatureHandler
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(
	cdc codec.Codec,
	signatureOrderHandlers ...bandtssclient.RequestSignatureHandler,
) AppModuleBasic {
	return AppModuleBasic{
		cdc:                    cdc,
		signatureOrderHandlers: signatureOrderHandlers,
	}
}

// Name returns the module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the bandtss module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the bandtss module's interface types
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns default genesis state as raw bytes for the restake module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the restake module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bandtss module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the transaction commands for the bandtss module.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	signatureOrderHandlers := getSignatureOrderCLIHandlers(a.signatureOrderHandlers)

	return cli.GetTxCmd(signatureOrderHandlers)
}

func getSignatureOrderCLIHandlers(handlers []bandtssclient.RequestSignatureHandler) []*cobra.Command {
	signatureOrderHandlers := make([]*cobra.Command, 0, len(handlers))
	for _, handler := range handlers {
		signatureOrderHandlers = append(signatureOrderHandlers, handler.CLIHandler())
	}
	return signatureOrderHandlers
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement.
type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(
	cdc codec.Codec,
	k *keeper.Keeper,
	signatureOrderHandlers ...bandtssclient.RequestSignatureHandler,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc, signatureOrderHandlers...),
		keeper:         k,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers a GRPC query service to respond to the module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	msgServer := keeper.NewMsgServerImpl(am.keeper)
	types.RegisterMsgServer(cfg.MsgServer(), msgServer)
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// RegisterInvariants registers the bandtss module's invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	am.keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// BeginBlock processes ABCI begin block message for this module (SDK AppModule interface).
func (am AppModule) BeginBlock(ctx context.Context) error {
	c := sdk.UnwrapSDKContext(ctx)
	return BeginBlocker(c, am.keeper)
}

// EndBlock processes ABCI end block message for the module (SDK AppModule interface).
func (am AppModule) EndBlock(ctx context.Context) error {
	c := sdk.UnwrapSDKContext(ctx)
	return EndBlocker(c, am.keeper)
}
