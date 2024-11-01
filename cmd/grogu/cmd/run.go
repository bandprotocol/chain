package cmd

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client"

	"github.com/bandprotocol/chain/v3/grogu/context"
	"github.com/bandprotocol/chain/v3/grogu/querier"
	"github.com/bandprotocol/chain/v3/grogu/signaller"
	"github.com/bandprotocol/chain/v3/grogu/submitter"
	"github.com/bandprotocol/chain/v3/grogu/updater"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

const (
	flagValidator            = "validator"
	flagNodes                = "nodes"
	flagBroadcastTimeout     = "broadcast-timeout"
	flagRPCPollInterval      = "rpc-poll-interval"
	flagMaxTry               = "max-try"
	flagBothan               = "bothan"
	flagBothanTimeout        = "bothan-timeout"
	flagDistrStartPct        = "distribution-start-pct"
	flagDistrOffsetPct       = "distribution-offset-pct"
	flagLogLevel             = "log-level"
	flagUpdaterQueryInterval = "updater-query-interval"
)

func RunCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the grogu process",
		Args:    cobra.ExactArgs(0),
		RunE:    createRunE(ctx),
	}

	cmd.Flags().String(flagValidator, "", "The validator address to send messages for.")
	cmd.Flags().String(flagNodes, "tcp://localhost:26657", "The RPC URLs to connect to.")
	cmd.Flags().String(flags.FlagChainID, "", "The chain ID of the connected chain.")
	cmd.Flags().String(flags.FlagGasPrices, "0uband", "The gas prices for transactions.")
	cmd.Flags().String(flagBroadcastTimeout, "1m", "The timeout duration for transaction commits.")
	cmd.Flags().String(flagRPCPollInterval, "1s", "The duration to wait between RPC polls.")
	cmd.Flags().Uint64(flagMaxTry, 5, "The maximum number of attempts to submit a transaction.")
	cmd.Flags().Uint64(flagDistrStartPct, 50, "The starting percentage for the distribution offset range.")
	cmd.Flags().Uint64(flagDistrOffsetPct, 30, "The offset percentage range from the starting distribution.")
	cmd.Flags().String(flagBothan, "", "The Bothan URL to connect to.")
	cmd.Flags().String(flagBothanTimeout, "10s", "The timeout duration for Bothan requests.")
	cmd.Flags().String(flagLogLevel, "info", "The application's log level.")
	cmd.Flags().String(flagUpdaterQueryInterval, "1m", "The interval for updater querying chain.")

	_ = viper.BindPFlag(flagValidator, cmd.Flags().Lookup(flagValidator))
	_ = viper.BindPFlag(flagNodes, cmd.Flags().Lookup(flagNodes))
	_ = viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))
	_ = viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))
	_ = viper.BindPFlag(flagBroadcastTimeout, cmd.Flags().Lookup(flagBroadcastTimeout))
	_ = viper.BindPFlag(flagRPCPollInterval, cmd.Flags().Lookup(flagRPCPollInterval))
	_ = viper.BindPFlag(flagMaxTry, cmd.Flags().Lookup(flagMaxTry))
	_ = viper.BindPFlag(flagDistrStartPct, cmd.Flags().Lookup(flagDistrStartPct))
	_ = viper.BindPFlag(flagDistrOffsetPct, cmd.Flags().Lookup(flagDistrOffsetPct))
	_ = viper.BindPFlag(flagBothan, cmd.Flags().Lookup(flagBothan))
	_ = viper.BindPFlag(flagBothanTimeout, cmd.Flags().Lookup(flagBothanTimeout))
	_ = viper.BindPFlag(flagLogLevel, cmd.Flags().Lookup(flagLogLevel))
	_ = viper.BindPFlag(flagUpdaterQueryInterval, cmd.Flags().Lookup(flagUpdaterQueryInterval))

	return cmd
}

func createRunE(ctx *context.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		allowLevel, err := log.ParseLogLevel(ctx.Config.LogLevel)
		if err != nil {
			return err
		}
		l := logger.NewLogger(allowLevel)

		// Split Node URIs and create RPC clients
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		clientCtx = clientCtx.WithKeyring(ctx.Keyring).
			WithChainID(viper.GetString(flags.FlagChainID)).
			WithCodec(ctx.BandApp.AppCodec()).
			WithInterfaceRegistry(ctx.BandApp.InterfaceRegistry()).
			WithTxConfig(ctx.BandApp.GetTxConfig()).
			WithBroadcastMode(flags.BroadcastSync)

		nodeURIs := strings.Split(viper.GetString(flagNodes), ",")
		clients, stopClients, err := createClients(nodeURIs)
		if err != nil {
			return err
		}
		defer stopClients()

		// Set up Queriers
		maxBlockHeight := new(atomic.Int64)
		maxBlockHeight.Store(0)

		authQuerier := querier.NewAuthQuerier(clientCtx, clients, maxBlockHeight)
		feedQuerier := querier.NewFeedQuerier(clientCtx, clients, maxBlockHeight)
		txQuerier := querier.NewTxQuerier(clientCtx, clients)

		// Setup Bothan service
		timeout, err := time.ParseDuration(ctx.Config.BothanTimeout)
		if err != nil {
			return err
		}
		bothanService, err := bothan.NewGrpcClient(ctx.Config.Bothan, timeout)
		if err != nil {
			return err
		}

		// Create submit channel
		submitSignalPriceCh := make(chan submitter.SignalPriceSubmission, 300)

		// Parse validator address
		valAddr, err := sdk.ValAddressFromBech32(ctx.Config.Validator)
		if err != nil {
			return err
		}

		// Parse broadcast timeout
		broadcastTimeout, err := time.ParseDuration(ctx.Config.BroadcastTimeout)
		if err != nil {
			return err
		}

		// Parse RPC poll interval
		rpcPollInterval, err := time.ParseDuration(ctx.Config.RPCPollInterval)
		if err != nil {
			return err
		}

		// Parse Updater query interval
		updaterQueryInterval, err := time.ParseDuration(ctx.Config.UpdaterQueryInterval)
		if err != nil {
			return err
		}

		// Initialize pending signal IDs map
		pendingSignalIDs := sync.Map{}

		// Setup Signaller
		signallerService := signaller.New(
			feedQuerier,
			bothanService,
			time.Second,
			submitSignalPriceCh,
			l,
			valAddr,
			&pendingSignalIDs,
			ctx.Config.DistributionStartPercentage,
			ctx.Config.DistributionOffsetPercentage,
		)

		// Setup Submitter
		submitterService, err := submitter.New(
			clientCtx,
			clients,
			bothanService,
			l,
			submitSignalPriceCh,
			authQuerier,
			txQuerier,
			valAddr,
			&pendingSignalIDs,
			broadcastTimeout,
			ctx.Config.MaxTry,
			rpcPollInterval,
			ctx.Config.GasPrices,
		)
		if err != nil {
			return err
		}

		// Setup Updater
		maxCurrentFeedEventHeight := new(atomic.Int64)
		maxCurrentFeedEventHeight.Store(0)

		maxUpdateRefSourceEventHeight := new(atomic.Int64)
		maxUpdateRefSourceEventHeight.Store(0)

		updaterService := updater.New(
			feedQuerier,
			bothanService,
			clients,
			l,
			updaterQueryInterval,
		)

		// Listen for termination signals for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start all services
		go updaterService.Start(sigChan)
		go signallerService.Start()
		go submitterService.Start()

		l.Info("Grogu has started")

		<-sigChan
		l.Info("Received stop signal, shutting down")
		l.Info("Grogu has stopped")

		return nil
	}
}
