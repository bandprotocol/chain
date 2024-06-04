package bandtss

import (
	"context"
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	bandtssclient "github.com/bandprotocol/chain/v2/x/bandtss/client"
	"github.com/bandprotocol/chain/v2/x/bandtss/client/cli"
	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.BeginBlockAppModule = AppModule{}
	_ module.EndBlockAppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the bandtss module.
type AppModuleBasic struct {
	signatureOrderHandlers []bandtssclient.RequestSignatureHandler
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(signatureOrderHandlers ...bandtssclient.RequestSignatureHandler) AppModuleBasic {
	return AppModuleBasic{
		signatureOrderHandlers: signatureOrderHandlers,
	}
}

// Name returns the bandtss module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// DefaultGenesis is an empty object.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis is always successful, as we ignore the value.
func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, config client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// GetQueryCmd returns the cli query commands for the bandtss module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
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

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bandtss module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the bandtss module's interface types
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// RegisterLegacyAminoCodec registers the bandtss module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement.
type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the bandtss module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// RegisterServices registers a GRPC query service to respond to the module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	msgServer := keeper.NewMsgServerImpl(am.keeper)
	types.RegisterMsgServer(cfg.MsgServer(), msgServer)
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// RegisterInvariants registers the bandtss module's invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs genesis initialization for the bandtss module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis is always empty, as InitGenesis does nothing either.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// BeginBlock processes ABCI begin block message for this bandtss module (SDK AppModule interface).
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	handleBeginBlock(ctx, req, am.keeper)
}

// EndBlock processes ABCI end block message for this bandtss module (SDK AppModule interface).
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	handleEndBlock(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
