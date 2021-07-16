package gov

// DONTCOVER

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	abci "github.com/tendermint/tendermint/abci/types"

	odingovkeeper "github.com/GeoDB-Limited/odin-core/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the gov module.
type AppModuleBasic struct {
	gov.AppModuleBasic
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(appModuleBasic gov.AppModuleBasic) AppModuleBasic {
	return AppModuleBasic{
		AppModuleBasic: appModuleBasic,
	}
}

//____________________________________________________________________________

// AppModule implements an application module for the gov module.
type AppModule struct {
	gov.AppModule

	keeper        odingovkeeper.Keeper
	accountKeeper govtypes.AccountKeeper
	bankKeeper    govtypes.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(appModule gov.AppModule, keeper odingovkeeper.Keeper, ak govtypes.AccountKeeper, bk govtypes.BankKeeper) AppModule {
	return AppModule{
		AppModule:     appModule,
		keeper:        keeper,
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

// LegacyQuerierHandler returns no sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return odingovkeeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	govtypes.RegisterMsgServer(cfg.MsgServer(), govkeeper.NewMsgServerImpl(am.keeper.Keeper))
	govtypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// EndBlock returns the end blocker for the gov module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}
