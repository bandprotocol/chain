package band

import (
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkparams "github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/bandtss"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	"github.com/bandprotocol/chain/v3/x/feeds"
	feedsclient "github.com/bandprotocol/chain/v3/x/feeds/client"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/globalfee"
	globalfeetypes "github.com/bandprotocol/chain/v3/x/globalfee/types"
	"github.com/bandprotocol/chain/v3/x/oracle"
	oracleclient "github.com/bandprotocol/chain/v3/x/oracle/client"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	"github.com/bandprotocol/chain/v3/x/restake"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	"github.com/bandprotocol/chain/v3/x/rollingseed"
	rollingseedtypes "github.com/bandprotocol/chain/v3/x/rollingseed/types"
	"github.com/bandprotocol/chain/v3/x/tss"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	"github.com/bandprotocol/chain/v3/x/tunnel"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	distrtypes.ModuleName:          nil,
	icatypes.ModuleName:            nil,
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:            {authtypes.Burner},
	ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	ibcfeetypes.ModuleName:         nil,
	bandtsstypes.ModuleName:        nil,
	restaketypes.ModuleName:        nil,
	tunneltypes.ModuleName:         {authtypes.Minter},
}

func appModules(
	app *BandApp,
	appCodec codec.Codec,
	txConfig client.TxEncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModule {
	return []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(
			appCodec,
			app.GovKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.GetSubspace(govtypes.ModuleName),
		),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(
			appCodec,
			app.SlashingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(slashingtypes.ModuleName),
			app.interfaceRegistry,
		),
		distr.NewAppModule(
			appCodec,
			app.DistrKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(distrtypes.ModuleName),
		),
		staking.NewAppModule(
			appCodec,
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.GetSubspace(stakingtypes.ModuleName),
		),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		feegrantmodule.NewAppModule(
			appCodec,
			app.AccountKeeper,
			app.BankKeeper,
			app.FeeGrantKeeper,
			app.interfaceRegistry,
		),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		ibctm.NewAppModule(),
		sdkparams.NewAppModule(app.ParamsKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		oracle.NewAppModule(
			appCodec,
			app.OracleKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(oracletypes.ModuleName),
		),
		app.TransferModule,
		app.ICAModule,
		globalfee.NewAppModule(app.GlobalFeeKeeper),
		ibcfee.NewAppModule(app.IBCFeeKeeper),
		restake.NewAppModule(appCodec, app.RestakeKeeper),
		feeds.NewAppModule(appCodec, app.FeedsKeeper),
		rollingseed.NewAppModule(appCodec, app.RollingseedKeeper),
		tss.NewAppModule(appCodec, app.TSSKeeper),
		bandtss.NewAppModule(
			appCodec,
			app.BandtssKeeper,
			oracleclient.OracleRequestSignatureHandler,
			feedsclient.FeedsRequestSignatureHandler,
		),
		tunnel.NewAppModule(appCodec, app.TunnelKeeper),
	}
}

// ModuleBasics defines the module BasicManager that is in charge of setting up basic,
// non-dependant module elements, such as codec registration
// and genesis verification.
func newBasicManagerFromManager(app *BandApp) module.BasicManager {
	basicManager := module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
				},
			),
		})
	basicManager.RegisterLegacyAminoCodec(app.legacyAmino)
	basicManager.RegisterInterfaces(app.interfaceRegistry)
	return basicManager
}

// simulationModules returns modules for simulation manager
// define the order of the modules for deterministic simulations
func simulationModules(
	app *BandApp,
	appCodec codec.Codec,
	_ bool,
) []module.AppModuleSimulation {
	return []module.AppModuleSimulation{
		auth.NewAppModule(
			appCodec,
			app.AccountKeeper,
			authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(
			appCodec,
			app.AccountKeeper,
			app.BankKeeper,
			app.FeeGrantKeeper,
			app.interfaceRegistry,
		),
		gov.NewAppModule(
			appCodec,
			app.GovKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.GetSubspace(govtypes.ModuleName),
		),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		staking.NewAppModule(
			appCodec,
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.GetSubspace(stakingtypes.ModuleName),
		),
		distr.NewAppModule(
			appCodec,
			app.DistrKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(distrtypes.ModuleName),
		),
		slashing.NewAppModule(
			appCodec,
			app.SlashingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(slashingtypes.ModuleName),
			app.interfaceRegistry,
		),
		sdkparams.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		oracle.NewAppModule(
			appCodec,
			app.OracleKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(oracletypes.ModuleName),
		),
		ibc.NewAppModule(app.IBCKeeper),
		app.TransferModule,
	}
}

func orderBeginBlockers() []string {
	return []string{
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		rollingseedtypes.ModuleName,
		oracletypes.ModuleName,
		tsstypes.ModuleName,
		bandtsstypes.ModuleName,
		restaketypes.ModuleName,
		feedstypes.ModuleName,
		tunneltypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		globalfeetypes.ModuleName,
	}
}

/*
Interchain Security Requirements:
- provider.EndBlock gets validator updates from the staking module;
thus, staking.EndBlock must be executed before provider.EndBlock;
- creating a new consumer chain requires the following order,
CreateChildClient(), staking.EndBlock, provider.EndBlock;
thus, gov.EndBlock must be executed before staking.EndBlock
*/
func orderEndBlockers() []string {
	return []string{
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		rollingseedtypes.ModuleName,
		oracletypes.ModuleName,
		tsstypes.ModuleName,
		bandtsstypes.ModuleName,
		restaketypes.ModuleName,
		feedstypes.ModuleName,
		tunneltypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		capabilitytypes.ModuleName,
		ibcfeetypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		globalfeetypes.ModuleName,
	}
}

/*
NOTE: The genutils module must occur after staking so that pools are
properly initialized with tokens from genesis accounts.
NOTE: The genutils module must also occur after auth so that it can access the params from auth.
NOTE: Capability module must occur first so that it can initialize any capabilities
so that other modules that want to create or claim capabilities afterwards in InitChain
can do so safely.
*/
func orderInitBlockers() []string {
	return []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		rollingseedtypes.ModuleName,
		oracletypes.ModuleName,
		tsstypes.ModuleName,
		bandtsstypes.ModuleName,
		globalfeetypes.ModuleName,
		restaketypes.ModuleName,
		feedstypes.ModuleName,
		tunneltypes.ModuleName,
	}
}
