package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingcli "github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/app/params"
	"github.com/bandprotocol/chain/v2/hooks/emitter"
	"github.com/bandprotocol/chain/v2/hooks/price"
	"github.com/bandprotocol/chain/v2/hooks/request"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	flagWithEmitter            = "with-emitter"
	flagDisableFeelessReports  = "disable-feeless-reports"
	flagEnableFastSync         = "enable-fast-sync"
	flagWithPricer             = "with-pricer"
	flagWithRequestSearch      = "with-request-search"
	flagRequestSearchCacheSize = "request-search-cache-size"
	flagWithOwasmCacheSize     = "oracle-script-cache-size"
)

// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := band.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(band.DefaultNodeHome).
		WithViper("BAND")

	rootCmd := &cobra.Command{
		Use:   "bandd",
		Short: "BandChain App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = client.ReadHomeFlag(initClientCtx, cmd)

			initClientCtx, err := config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd, "", nil)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	rootCmd.AddCommand(
		InitCmd(band.NewDefaultGenesisState(), band.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, band.DefaultNodeHome),
		band.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(band.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, band.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(band.ModuleBasics),
		AddGenesisAccountCmd(band.DefaultNodeHome),
		AddGenesisDataSourceCmd(band.DefaultNodeHome),
		AddGenesisOracleScriptCmd(band.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		// testnetCmd(band.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		config.Cmd(),
	)

	ac := appCreator{
		encCfg: encodingConfig,
	}
	server.AddCommands(rootCmd, band.DefaultNodeHome, ac.newApp, ac.appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(band.DefaultNodeHome),
	)
}
func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	startCmd.Flags().Uint32(flagWithOwasmCacheSize, 100, "[Experimental] Number of oracle scripts to cache")
	startCmd.Flags().Bool(flagDisableFeelessReports, false, "Disable feeless reports during congestion")
	startCmd.Flags().String(flagWithRequestSearch, "", "[Experimental] Enable mode to save request in sql database")
	startCmd.Flags().Int(flagRequestSearchCacheSize, 10, "[Experimental] indicates number of latest oracle requests to be stored in database")
	startCmd.Flags().String(flagWithEmitter, "", "[Experimental] Enable mode to save request in sql database")
	startCmd.Flags().String(flagWithPricer, "", "[Experimental] Enable collecting standard price reference provided by given oracle scripts and save in level db (ex. ids/defaultAskCount/defaultMinCount)")
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subandommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	band.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subandommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		MultiSendTxCmd(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
		vestingcli.GetTxCmd(),
	)

	band.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

type appCreator struct {
	encCfg params.EncodingConfig
}

// newApp is an AppCreator
func (ac appCreator) newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	bandApp := band.NewBandApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		ac.encCfg,
		appOpts,
		cast.ToBool(appOpts.Get(flagDisableFeelessReports)),
		cast.ToUint32(appOpts.Get(flagWithOwasmCacheSize)),
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
	connStr, _ := appOpts.Get(flagWithRequestSearch).(string)
	if connStr != "" {
		requestSearchCacheSize := appOpts.Get(flagRequestSearchCacheSize).(int)
		bandApp.AddHook(request.NewHook(
			bandApp.AppCodec(), bandApp.OracleKeeper, connStr, requestSearchCacheSize))
	}

	connStr, _ = appOpts.Get(flagWithEmitter).(string)
	if connStr != "" {
		bandApp.AddHook(
			emitter.NewHook(bandApp.AppCodec(), bandApp.LegacyAmino(), band.MakeEncodingConfig(), bandApp.AccountKeeper, bandApp.BankKeeper,
				bandApp.StakingKeeper, bandApp.MintKeeper, bandApp.DistrKeeper, bandApp.GovKeeper,
				bandApp.OracleKeeper, bandApp.IBCKeeper.ClientKeeper, bandApp.IBCKeeper.ConnectionKeeper, bandApp.IBCKeeper.ChannelKeeper, connStr, false))
	}

	pricerStr, _ := appOpts.Get(flagWithPricer).(string)
	if pricerStr != "" {
		pricerStrArgs := strings.Split(pricerStr, "/")
		var defaultAskCount, defaultMinCount uint64
		if len(pricerStrArgs) == 3 {
			defaultAskCount, err = strconv.ParseUint(pricerStrArgs[1], 10, 64)
			if err != nil {
				panic(err)
			}
			defaultMinCount, err = strconv.ParseUint(pricerStrArgs[2], 10, 64)
			if err != nil {
				panic(err)
			}
		} else if len(pricerStrArgs) == 2 || len(pricerStrArgs) > 3 {
			panic(fmt.Errorf("accepts 1 or 3 arg(s), received %d", len(pricerStrArgs)))
		}
		rawOracleIDs := strings.Split(pricerStrArgs[0], ",")
		var oracleIDs []oracletypes.OracleScriptID
		for _, rawOracleID := range rawOracleIDs {
			oracleID, err := strconv.ParseInt(rawOracleID, 10, 64)
			if err != nil {
				panic(err)
			}
			oracleIDs = append(oracleIDs, oracletypes.OracleScriptID(oracleID))
		}
		bandApp.AddHook(
			price.NewHook(bandApp.AppCodec(), bandApp.OracleKeeper, oracleIDs,
				filepath.Join(cast.ToString(appOpts.Get(cli.HomeFlag)), "prices"),
				defaultAskCount, defaultMinCount))
	}

	return bandApp
}

func (ac appCreator) appExport(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string,
	appOpts servertypes.AppOptions) (servertypes.ExportedApp, error) {

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	var loadLatest bool
	if height == -1 {
		loadLatest = true
	}

	bandApp := band.NewBandApp(
		logger,
		db,
		traceStore,
		loadLatest,
		map[int64]bool{},
		homePath,
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		ac.encCfg,
		appOpts,
		false,
		cast.ToUint32(appOpts.Get(flagWithOwasmCacheSize)),
	)

	if height != -1 {
		if err := bandApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return bandApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}
