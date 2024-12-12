package keepers

import (
	"os"
	"path/filepath"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	owasm "github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v3/x/bandtss"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	bandbankkeeper "github.com/bandprotocol/chain/v3/x/bank/keeper"
	"github.com/bandprotocol/chain/v3/x/feeds"
	feedskeeper "github.com/bandprotocol/chain/v3/x/feeds/keeper"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	globalfeekeeper "github.com/bandprotocol/chain/v3/x/globalfee/keeper"
	globalfeetypes "github.com/bandprotocol/chain/v3/x/globalfee/types"
	"github.com/bandprotocol/chain/v3/x/oracle"
	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restakekeeper "github.com/bandprotocol/chain/v3/x/restake/keeper"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	rollingseedkeeper "github.com/bandprotocol/chain/v3/x/rollingseed/keeper"
	rollingseedtypes "github.com/bandprotocol/chain/v3/x/rollingseed/types"
	"github.com/bandprotocol/chain/v3/x/tss"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	"github.com/bandprotocol/chain/v3/x/tunnel"
	tunnelkeeper "github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	RollingseedKeeper     rollingseedkeeper.Keeper
	OracleKeeper          oraclekeeper.Keeper
	TSSKeeper             *tsskeeper.Keeper
	BandtssKeeper         bandtsskeeper.Keeper
	FeedsKeeper           feedskeeper.Keeper
	TunnelKeeper          tunnelkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	GlobalFeeKeeper       globalfeekeeper.Keeper
	RestakeKeeper         restakekeeper.Keeper
	// IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCKeeper      *ibckeeper.Keeper
	ICAHostKeeper  icahostkeeper.Keeper
	IBCFeeKeeper   ibcfeekeeper.Keeper
	TransferKeeper ibctransferkeeper.Keeper

	// Modules
	ICAModule      ica.AppModule
	TransferModule transfer.AppModule

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedOracleKeeper   capabilitykeeper.ScopedKeeper
	ScopedTunnelKeeper   capabilitykeeper.ScopedKeeper
}

func NewAppKeeper(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	legacyAmino *codec.LegacyAmino,
	maccPerms map[string][]string,
	modAccAddrs map[string]bool,
	blockedAddress map[string]bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	logger log.Logger,
	appOpts servertypes.AppOptions,
	owasmCacheSize uint32,
) AppKeepers {
	appKeepers := AppKeepers{}

	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	appKeepers.GenerateKeys()

	/*
		configure state listening capabilities using AppOptions
		we are doing nothing with the returned streamingServices and waitGroup in this case
	*/
	// load state streaming if enabled

	if err := bApp.RegisterStreamingServices(appOpts, appKeepers.keys); err != nil {
		logger.Error("failed to load state streaming", "err", err)
		os.Exit(1)
	}

	appKeepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		appKeepers.keys[paramstypes.StoreKey],
		appKeepers.tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)

	bApp.SetParamStore(appKeepers.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[capabilitytypes.StoreKey],
		appKeepers.memKeys[capabilitytypes.MemStoreKey],
	)

	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	appKeepers.ScopedICAHostKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.ScopedOracleKeeper = appKeepers.CapabilityKeeper.ScopeToModule(oracletypes.ModuleName)
	appKeepers.ScopedTunnelKeeper = appKeepers.CapabilityKeeper.ScopeToModule(tunneltypes.ModuleName)

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	appKeepers.CapabilityKeeper.Seal()

	// Add normal keepers
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// wrappedBankerKeeper overrides burn token behavior to instead transfer to community pool.
	appKeepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[banktypes.StoreKey]),
		appKeepers.AccountKeeper,
		blockedAddress,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	wrapBank := bandbankkeeper.NewWrappedBankKeeperBurnToCommunityPool(
		appKeepers.BankKeeper,
		appKeepers.AccountKeeper,
		logger,
	)

	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[crisistypes.StoreKey]),
		invCheckPeriod,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		appKeepers.AccountKeeper.AddressCodec(),
	)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(appKeepers.keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[feegrant.StoreKey]),
		appKeepers.AccountKeeper,
	)

	// Using pointer of WrapBankKeeper in staking module
	appKeepers.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[stakingtypes.StoreKey]),
		appKeepers.AccountKeeper,
		wrapBank,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[minttypes.StoreKey]),
		appKeepers.StakingKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[distrtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	wrapBank.SetDistrKeeper(&appKeepers.DistrKeeper)

	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(appKeepers.keys[slashingtypes.StoreKey]),
		appKeepers.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.RestakeKeeper = restakekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[restaketypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks(),
			appKeepers.RestakeKeeper.Hooks(),
		),
	)

	// UpgradeKeeper must be created before IBCKeeper
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(appKeepers.keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// UpgradeKeeper must be created before IBCKeeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcexported.StoreKey],
		appKeepers.GetSubspace(ibcexported.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		appKeepers.ScopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	govConfig := govtypes.DefaultConfig()
	// set the MaxMetadataLen for proposals to the same value as it was pre-sdk v0.47.x
	govConfig.MaxMetadataLen = 10200
	appKeepers.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[govtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		bApp.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper))

	// Set legacy router for backwards compatibility with gov v1beta1
	appKeepers.GovKeeper.SetLegacyRouter(govRouter)

	appKeepers.GovKeeper = appKeepers.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[evidencetypes.StoreKey]),
		appKeepers.StakingKeeper,
		appKeepers.SlashingKeeper,
		appKeepers.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)

	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper
	// GlobalFeeKeeper
	appKeepers.GlobalFeeKeeper = globalfeekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[globalfeetypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec, appKeepers.keys[ibcfeetypes.StoreKey],
		appKeepers.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper, appKeepers.AccountKeeper, appKeepers.BankKeeper,
	)

	// ICA Host keeper
	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icahosttypes.StoreKey],
		appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.ScopedICAHostKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// required since ibc-go v7.5.0
	appKeepers.ICAHostKeeper.WithQueryRouter(bApp.GRPCQueryRouter())

	appKeepers.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		appKeepers.IBCFeeKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	owasmVM, err := owasm.NewVm(owasmCacheSize)
	if err != nil {
		panic(err)
	}

	appKeepers.RollingseedKeeper = rollingseedkeeper.NewKeeper(appKeepers.keys[rollingseedtypes.StoreKey])

	// register the request signature types
	tssContentRouter := tsstypes.NewContentRouter()
	tssCbRouter := tsstypes.NewCallbackRouter()

	appKeepers.TSSKeeper = tsskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[tsstypes.StoreKey],
		appKeepers.AuthzKeeper,
		appKeepers.RollingseedKeeper,
		tssContentRouter,
		tssCbRouter,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.BandtssKeeper = bandtsskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[bandtsstypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		appKeepers.TSSKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authtypes.FeeCollectorName,
	)

	appKeepers.OracleKeeper = oraclekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[oracletypes.StoreKey],
		filepath.Join(homePath, "files"),
		authtypes.FeeCollectorName,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		appKeepers.AuthzKeeper,
		appKeepers.IBCFeeKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.RollingseedKeeper,
		appKeepers.BandtssKeeper,
		appKeepers.ScopedOracleKeeper,
		owasmVM,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.FeedsKeeper = feedskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[feedstypes.StoreKey],
		appKeepers.OracleKeeper,
		appKeepers.StakingKeeper,
		appKeepers.RestakeKeeper,
		appKeepers.AuthzKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.TunnelKeeper = tunnelkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[tunneltypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.FeedsKeeper,
		appKeepers.BandtssKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCFeeKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedTunnelKeeper,
		appKeepers.TransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Add TSS route
	tssContentRouter.
		AddRoute(tsstypes.RouterKey, tss.NewSignatureOrderHandler(*appKeepers.TSSKeeper)).
		AddRoute(oracletypes.RouterKey, oracle.NewSignatureOrderHandler(appKeepers.OracleKeeper)).
		AddRoute(bandtsstypes.RouterKey, bandtss.NewSignatureOrderHandler()).
		AddRoute(feedstypes.RouterKey, feeds.NewSignatureOrderHandler(appKeepers.FeedsKeeper)).
		AddRoute(tunneltypes.RouterKey, tunnel.NewSignatureOrderHandler(appKeepers.TunnelKeeper))

	tssCbRouter.
		AddRoute(bandtsstypes.RouterKey, bandtsskeeper.NewTSSCallback(appKeepers.BandtssKeeper))

	// It is vital to seal the request signature router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	tssContentRouter.Seal()
	tssCbRouter.Seal()

	// Middleware Stacks
	appKeepers.ICAModule = ica.NewAppModule(nil, &appKeepers.ICAHostKeeper)
	appKeepers.TransferModule = transfer.NewAppModule(appKeepers.TransferKeeper)

	// Create Transfer Stack
	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(appKeepers.TransferKeeper)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, appKeepers.IBCFeeKeeper)

	// Create ICAHost Stack
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(appKeepers.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, appKeepers.IBCFeeKeeper)

	// Create Oracle Stack
	var oracleStack porttypes.IBCModule = oracle.NewIBCModule(appKeepers.OracleKeeper)

	// Create Tunnel Stack
	var tunnelStack porttypes.IBCModule = tunnel.NewIBCModule(appKeepers.TunnelKeeper)

	ibcRouter := porttypes.NewRouter().AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(oracletypes.ModuleName, oracleStack).
		AddRoute(tunneltypes.ModuleName, tunnelStack)

	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	return appKeepers
}

// GetSubspace returns a param subspace for a given module name.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, ok := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	if !ok {
		panic("couldn't load subspace for module: " + moduleName)
	}
	return subspace
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(
	appCodec codec.BinaryCodec,
	legacyAmino *codec.LegacyAmino,
	key, tkey storetypes.StoreKey,
) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable()) //nolint: staticcheck
	paramsKeeper.Subspace(stakingtypes.ModuleName).
		WithKeyTable(stakingtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(banktypes.ModuleName).
		WithKeyTable(banktypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(minttypes.ModuleName).
		WithKeyTable(minttypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(distrtypes.ModuleName).
		WithKeyTable(distrtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(slashingtypes.ModuleName).
		WithKeyTable(slashingtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(govtypes.ModuleName).
		WithKeyTable(govv1.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(crisistypes.ModuleName).
		WithKeyTable(crisistypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	paramsKeeper.Subspace(oracletypes.ModuleName)

	return paramsKeeper
}
