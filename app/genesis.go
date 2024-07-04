package band

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icagenesistypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransafertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/bandprotocol/chain/v2/app/upgrades/v2_6"
	"github.com/bandprotocol/chain/v2/x/bandtss"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/feeds"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	globalfeetypes "github.com/bandprotocol/chain/v2/x/globalfee/types"
	"github.com/bandprotocol/chain/v2/x/oracle"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	"github.com/bandprotocol/chain/v2/x/rollingseed"
	rollingseedtypes "github.com/bandprotocol/chain/v2/x/rollingseed/types"
	"github.com/bandprotocol/chain/v2/x/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	tunnel "github.com/bandprotocol/chain/v2/x/tunnel"
	tunneltypes "github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// GenesisState defines a type alias for the Band genesis application state.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	cdc := MakeEncodingConfig().Marshaler
	ModuleBasics.DefaultGenesis(cdc)
	denom := "uband"
	// Get default genesis states of the modules we are to override.
	authGenesis := authtypes.DefaultGenesisState()
	stakingGenesis := stakingtypes.DefaultGenesisState()
	distrGenesis := distrtypes.DefaultGenesisState()
	mintGenesis := minttypes.DefaultGenesisState()
	govGenesis := govv1.DefaultGenesisState()
	crisisGenesis := crisistypes.DefaultGenesisState()
	slashingGenesis := slashingtypes.DefaultGenesisState()
	icaGenesis := icagenesistypes.DefaultGenesis()
	globalfeeGenesis := globalfeetypes.DefaultGenesisState()
	// Override the genesis parameters.
	authGenesis.Params.TxSizeCostPerByte = 5
	stakingGenesis.Params.BondDenom = denom
	stakingGenesis.Params.HistoricalEntries = 1000
	mintGenesis.Params.BlocksPerYear = 10519200 // target 3-second block time
	mintGenesis.Params.MintDenom = denom
	govGenesis.Params.MinDeposit = sdk.NewCoins(
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction)),
	)
	crisisGenesis.ConstantFee = sdk.NewCoin(denom, sdk.TokensFromConsensusPower(10000, sdk.DefaultPowerReduction))
	slashingGenesis.Params.SignedBlocksWindow = 30000                         // approximately 1 day
	slashingGenesis.Params.MinSignedPerWindow = sdk.NewDecWithPrec(5, 2)      // 5%
	slashingGenesis.Params.DowntimeJailDuration = 60 * 10 * time.Second       // 10 minutes
	slashingGenesis.Params.SlashFractionDoubleSign = sdk.NewDecWithPrec(5, 2) // 5%
	slashingGenesis.Params.SlashFractionDowntime = sdk.NewDecWithPrec(1, 4)   // 0.01%

	icaGenesis.HostGenesisState.Params = icahosttypes.Params{
		HostEnabled:   true,
		AllowMessages: v2_6.ICAAllowMessages,
	}

	globalfeeGenesis.Params.MinimumGasPrices = sdk.NewDecCoins(
		sdk.NewDecCoinFromDec(denom, sdk.NewDecWithPrec(25, 4)),
	)

	return GenesisState{
		authtypes.ModuleName:         cdc.MustMarshalJSON(authGenesis),
		genutiltypes.ModuleName:      genutil.AppModuleBasic{}.DefaultGenesis(cdc),
		banktypes.ModuleName:         bank.AppModuleBasic{}.DefaultGenesis(cdc),
		capabilitytypes.ModuleName:   capability.AppModuleBasic{}.DefaultGenesis(cdc),
		stakingtypes.ModuleName:      cdc.MustMarshalJSON(stakingGenesis),
		minttypes.ModuleName:         cdc.MustMarshalJSON(mintGenesis),
		distrtypes.ModuleName:        cdc.MustMarshalJSON(distrGenesis),
		govtypes.ModuleName:          cdc.MustMarshalJSON(govGenesis),
		crisistypes.ModuleName:       cdc.MustMarshalJSON(crisisGenesis),
		slashingtypes.ModuleName:     cdc.MustMarshalJSON(slashingGenesis),
		ibcexported.ModuleName:       ibc.AppModuleBasic{}.DefaultGenesis(cdc),
		upgradetypes.ModuleName:      upgrade.AppModuleBasic{}.DefaultGenesis(cdc),
		evidencetypes.ModuleName:     evidence.AppModuleBasic{}.DefaultGenesis(cdc),
		authz.ModuleName:             authzmodule.AppModuleBasic{}.DefaultGenesis(cdc),
		feegrant.ModuleName:          feegrantmodule.AppModuleBasic{}.DefaultGenesis(cdc),
		group.ModuleName:             groupmodule.AppModuleBasic{}.DefaultGenesis(cdc),
		ibctransafertypes.ModuleName: ibctransfer.AppModuleBasic{}.DefaultGenesis(cdc),
		icatypes.ModuleName:          cdc.MustMarshalJSON(icaGenesis),
		rollingseedtypes.ModuleName:  rollingseed.AppModuleBasic{}.DefaultGenesis(cdc),
		oracletypes.ModuleName:       oracle.AppModuleBasic{}.DefaultGenesis(cdc),
		tsstypes.ModuleName:          tss.AppModuleBasic{}.DefaultGenesis(cdc),
		bandtsstypes.ModuleName:      bandtss.AppModuleBasic{}.DefaultGenesis(cdc),
		feedstypes.ModuleName:        feeds.AppModuleBasic{}.DefaultGenesis(cdc),
		tunneltypes.ModuleName:       tunnel.AppModuleBasic{}.DefaultGenesis(cdc),
		globalfeetypes.ModuleName:    cdc.MustMarshalJSON(globalfeeGenesis),
	}
}
