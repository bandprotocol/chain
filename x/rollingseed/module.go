package rollingseed

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/x/rollingseed/keeper"
	"github.com/bandprotocol/chain/v2/x/rollingseed/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the rollingseed module.
type AppModuleBasic struct{}

// Name returns the rollingseed module's name.
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

// GetQueryCmd returns the cli query commands for the rollingseed module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// GetTxCmd returns the transaction commands for the rollingseed module.
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the rollingseed module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
}

// RegisterInterfaces registers the rollingseed module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
}

// RegisterLegacyAminoCodec registers the rollingseed module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
}

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement.
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the rollingseed module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// Deprecated: Route returns the message routing key for the rollingseed module.
func (am AppModule) Route() sdk.Route { return sdk.Route{} }

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns the rollingseed module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// RegisterInvariants registers the rollingseed module's invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs a no-op.
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

// BeginBlock processes ABCI begin block message for this rollingseed module (SDK AppModule interface).
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	handleBeginBlock(ctx, req, am.keeper)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
