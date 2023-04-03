package bank

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModule           = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the distribution module.
type AppModuleBasic struct {
	bank.AppModuleBasic
}

// AppModule implements an application module for the bank module.
type AppModule struct {
	bank.AppModule

	keeper        bankkeeper.Keeper
	accountKeeper types.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper bankkeeper.Keeper, accountKeeper types.AccountKeeper) AppModule {
	return AppModule{
		AppModule:     bank.NewAppModule(cdc, keeper, accountKeeper),
		keeper:        keeper,
		accountKeeper: accountKeeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), bankkeeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := bankkeeper.NewMigrator(am.keeper.(keeper.WrappedBankKeeper).Keeper.(bankkeeper.BaseKeeper))
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 2 to 3: %v", err))
	}
}
