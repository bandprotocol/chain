package auction

import (
	"context"
	"encoding/json"
	auctioncli "github.com/GeoDB-Limited/odin-core/x/auction/client/cli"
	auctionrest "github.com/GeoDB-Limited/odin-core/x/auction/client/rest"
	auctionkeeper "github.com/GeoDB-Limited/odin-core/x/auction/keeper"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
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

// Name returns this module's name - "auction" (SDK AppModuleBasic interface).
func (AppModuleBasic) Name() string {
	return auctiontypes.ModuleName
}

// RegisterLegacyAminoCodec registers the staking module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	auctiontypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	auctiontypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns the default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(auctiontypes.DefaultGenesisState())
}

// ValidateGenesis checks of the Genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs auctiontypes.GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		return err
	}
	return auctiontypes.ValidateGenesis(gs)
}

// RegisterRESTRoutes adds oracle REST endpoints to the main mux (SDK AppModuleBasic interface).
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	auctionrest.RegisterRoutes(clientCtx, rtr)
	return
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the oracle module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	auctiontypes.RegisterQueryHandlerClient(context.Background(), mux, auctiontypes.NewQueryClient(clientCtx))
}

// GetTxCmd returns cobra CLI command to send txs for this module (SDK AppModuleBasic interface).
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return auctioncli.GetTxCmd()
}

// GetQueryCmd returns cobra CLI command to query chain state (SDK AppModuleBasic interface).
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return auctioncli.GetQueryCmd()
}

// AppModule represents the AppModule for this module.
type AppModule struct {
	AppModuleBasic
	keeper auctionkeeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(k auctionkeeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// RegisterInvariants is a noop function to satisfy SDK AppModule interface.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the module's path for message route (SDK AppModule interface).
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(auctiontypes.RouterKey, NewHandler(am.keeper))
}

// QuerierRoute returns the oracle module's querier route name.
func (AppModule) QuerierRoute() string {
	return auctiontypes.QuerierRoute
}

// LegacyQuerierHandler returns the staking module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return auctionkeeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	auctiontypes.RegisterMsgServer(cfg.MsgServer(), auctionkeeper.NewMsgServerImpl(am.keeper))
	auctiontypes.RegisterQueryServer(cfg.QueryServer(), auctionkeeper.Querier{Keeper: am.keeper})
}

// BeginBlock processes ABCI begin block message for this oracle module (SDK AppModule interface).
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	handleBeginBlock(ctx, am.keeper)
}

// EndBlock processes ABCI end block message for this oracle module (SDK AppModule interface).
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// InitGenesis performs genesis initialization for the oracle module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState auctiontypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	auctionkeeper.InitGenesis(ctx, am.keeper, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the current state as genesis raw bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := auctionkeeper.ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}
