package band

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctransfer "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer"
	ibctransafertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/core"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/bandprotocol/chain/v2/x/oracle"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

// GenesisState defines a type alias for the Band genesis application state.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	cdc := MakeEncodingConfig().Marshaler
	denom := "uband"
	// Get default genesis states of the modules we are to override.
	authGenesis := authtypes.DefaultGenesisState()
	stakingGenesis := stakingtypes.DefaultGenesisState()
	distrGenesis := distrtypes.DefaultGenesisState()
	mintGenesis := minttypes.DefaultGenesisState()
	govGenesis := govtypes.DefaultGenesisState()
	crisisGenesis := crisistypes.DefaultGenesisState()
	slashingGenesis := slashingtypes.DefaultGenesisState()
	// Override the genesis parameters.
	authGenesis.Params.TxSizeCostPerByte = 5
	stakingGenesis.Params.BondDenom = denom
	stakingGenesis.Params.HistoricalEntries = 1000
	distrGenesis.Params.BaseProposerReward = sdk.NewDecWithPrec(3, 2)   // 3%
	distrGenesis.Params.BonusProposerReward = sdk.NewDecWithPrec(12, 2) // 12%
	mintGenesis.Params.BlocksPerYear = 10519200                         // target 3-second block time
	mintGenesis.Params.MintDenom = denom
	govGenesis.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(denom, sdk.TokensFromConsensusPower(1000)))
	crisisGenesis.ConstantFee = sdk.NewCoin(denom, sdk.TokensFromConsensusPower(10000))
	slashingGenesis.Params.SignedBlocksWindow = 30000                         // approximately 1 day
	slashingGenesis.Params.MinSignedPerWindow = sdk.NewDecWithPrec(5, 2)      // 5%
	slashingGenesis.Params.DowntimeJailDuration = 60 * 10 * time.Second       // 10 minutes
	slashingGenesis.Params.SlashFractionDoubleSign = sdk.NewDecWithPrec(5, 2) // 5%
	slashingGenesis.Params.SlashFractionDowntime = sdk.NewDecWithPrec(1, 4)   // 0.01%
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
		ibchost.ModuleName:           ibc.AppModuleBasic{}.DefaultGenesis(cdc),
		upgradetypes.ModuleName:      upgrade.AppModuleBasic{}.DefaultGenesis(cdc),
		evidencetypes.ModuleName:     evidence.AppModuleBasic{}.DefaultGenesis(cdc),
		ibctransafertypes.ModuleName: ibctransfer.AppModuleBasic{}.DefaultGenesis(cdc),
		oracletypes.ModuleName:       oracle.AppModuleBasic{}.DefaultGenesis(cdc),
	}
}
