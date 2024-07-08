package cmd

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/grogu/querier"
	"github.com/bandprotocol/chain/v2/grogu/signaller"
	"github.com/bandprotocol/chain/v2/grogu/submitter"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	flagValidator        = "validator"
	flagNodes            = "nodes"
	flagBroadcastTimeout = "broadcast-timeout"
	flagRPCPollInterval  = "rpc-poll-interval"
	flagMaxTry           = "max-try"
	flagBothan           = "bothan"
	flagBothanTimeout    = "bothan-timeout"
	flagDistrStartPct    = "distribution-start-pct"
	flagDistrOffsetPct   = "distribution-offset-pct"
	flagLogLevel         = "log-level"
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

	return cmd
}

func createRunE(ctx *context.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Initialize encoding config
		bandConfig := band.MakeEncodingConfig()
		chainID := viper.GetString(flags.FlagChainID)

		// Split Node URIs and create RPC clients
		nodesURIs := strings.Split(viper.GetString(flagNodes), ",")
		clients := make([]client.Context, 0, len(nodesURIs))

		for _, URI := range nodesURIs {
			httpClient, err := http.New(URI, "/websocket")
			if err != nil {
				return err
			}
			cl := client.Context{
				Client:            httpClient,
				ChainID:           chainID,
				Codec:             bandConfig.Marshaler,
				InterfaceRegistry: bandConfig.InterfaceRegistry,
				Keyring:           ctx.Keyring,
				TxConfig:          bandConfig.TxConfig,
				BroadcastMode:     flags.BroadcastSync,
			}
			clients = append(clients, cl)
		}

		// set up Queriers
		authQuerier := querier.NewAuthQuerier(clients)
		feedQuerier := querier.NewFeedQuerier(clients)
		txQuerier := querier.NewTxQuerier(clients)

		// Initialize logger
		allowLevel, err := log.AllowLevel(ctx.Config.LogLevel)
		if err != nil {
			return err
		}
		l := logger.New(allowLevel)

		// Setup Bothan service
		timeout, err := time.ParseDuration(ctx.Config.BothanTimeout)
		if err != nil {
			return err
		}
		bothanService, err := bothan.NewGRPC(ctx.Config.Bothan, timeout)
		if err != nil {
			return err
		}

		// Create submit channel
		submitSignalPriceCh := make(chan []types.SignalPrice, 300)

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
			clients,
			l,
			ctx.Keyring,
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

		// Start all
		go signallerService.Start()
		go submitterService.Start()

		l.Info("Grogu has started")

		// Listen for termination signals for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		<-sigChan
		l.Info("Received stop signal, shutting down")
		l.Info("Grogu has stopped")
		return nil
	}
}