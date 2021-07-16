package telemetry

import (
	"context"
	"encoding/json"
	telemetrycli "github.com/GeoDB-Limited/odin-core/x/telemetry/client/cli"
	telemetryrest "github.com/GeoDB-Limited/odin-core/x/telemetry/client/rest"
	telemetrykeeper "github.com/GeoDB-Limited/odin-core/x/telemetry/keeper"
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic is Band Oracle's module basic object.
type AppModuleBasic struct{}

// Name returns this module's name - "telemetry" (SDK AppModuleBasic interface).
func (AppModuleBasic) Name() string {
	return telemetrytypes.ModuleName
}

// RegisterLegacyAminoCodec registers the staking module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	telemetrytypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	telemetrytypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns the default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return []byte("{}")
}

// ValidateGenesis checks of the Genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes adds oracle REST endpoints to the main mux (SDK AppModuleBasic interface).
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	telemetryrest.RegisterRoutes(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the oracle module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	telemetrytypes.RegisterQueryHandlerClient(context.Background(), mux, telemetrytypes.NewQueryClient(clientCtx))
}

// GetTxCmd returns cobra CLI command to send txs for this module (SDK AppModuleBasic interface).
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns cobra CLI command to query chain state (SDK AppModuleBasic interface).
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return telemetrycli.GetQueryCmd()
}

// AppModule represents the AppModule for this module.
type AppModule struct {
	AppModuleBasic
	keeper telemetrykeeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(k telemetrykeeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// RegisterInvariants is a noop function to satisfy SDK AppModule interface.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the module's path for message route (SDK AppModule interface).
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the oracle module's querier route name.
func (AppModule) QuerierRoute() string {
	return telemetrytypes.QuerierRoute
}

// LegacyQuerierHandler returns the staking module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return telemetrykeeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	telemetrytypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// BeginBlock processes ABCI begin block message for this oracle module (SDK AppModule interface).
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
}

// EndBlock processes ABCI end block message for this oracle module (SDK AppModule interface).
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// InitGenesis performs genesis initialization for the oracle module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the current state as genesis raw bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return am.DefaultGenesis(cdc)
}
