package band

import (
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
)

// TestingApp functions

// GetBaseApp implements the TestingApp interface.
func (app *BandApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetTxConfig implements the TestingApp interface.
func (app *BandApp) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// GetTestGovKeeper implements the TestingApp interface.
func (app *BandApp) GetTestGovKeeper() *govkeeper.Keeper {
	return app.AppKeepers.GovKeeper
}

// GetStakingKeeper implements the TestingApp interface. Needed for ICS.
func (app *BandApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *BandApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *BandApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}
