package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/cylinder/workers/de"
	"github.com/bandprotocol/chain/v3/cylinder/workers/group"
	"github.com/bandprotocol/chain/v3/cylinder/workers/sender"
	"github.com/bandprotocol/chain/v3/cylinder/workers/signing"
)

const (
	flagGranter            = "granter"
	flagLogLevel           = "log-level"
	flagMaxMessages        = "max-messages"
	flagBroadcastTimeout   = "broadcast-timeout"
	flagRPCPollInterval    = "rpc-poll-interval"
	flagMaxTry             = "max-try"
	flagMinDE              = "min-de"
	flagGasAdjustStart     = "gas-adjust-start"
	flagGasAdjustStep      = "gas-adjust-step"
	flagRandomSecret       = "random-secret"
	flagCheckingDEInterval = "checking-de-interval"
	flagMetricsListenAddr  = "metrics-listen-addr"
)

// runCmd returns a Cobra command to run the cylinder process.
func runCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the cylinder process",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			group, err := group.New(ctx)
			if err != nil {
				return err
			}

			de, err := de.New(ctx)
			if err != nil {
				return err
			}

			signing, err := signing.New(ctx)
			if err != nil {
				return err
			}

			sender, err := sender.New(ctx)
			if err != nil {
				return err
			}

			workers := cylinder.Workers{group, de, signing, sender}

			// Start metrics server if enabled
			if ctx.Config.MetricsListenAddr != "" {
				ctx.Logger.Info("Metrics server is enabled")
				err := metrics.StartServer(cmd.Context(), ctx.Logger, ctx.Config.MetricsListenAddr)
				if err != nil {
					return fmt.Errorf("failed to start metrics server: %w", err)
				}

				// Add metrics from import data
				groups, err := ctx.Store.GetAllGroups()
				if err != nil {
					return err
				}
				metrics.AddGroupCount(float64(len(groups)))
				dkgs, err := ctx.Store.GetAllDKGs()
				if err != nil {
					return err
				}
				metrics.AddDKGLeftGauge(float64(len(dkgs)))
				des, err := ctx.Store.GetAllDEs()
				if err != nil {
					return err
				}
				metrics.AddOffChainDELeftGauge(float64(len(des)))
			}

			return cylinder.Run(ctx, workers)
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "chain ID of BandChain network")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "RPC url to BandChain node")
	cmd.Flags().String(flagGranter, "", "granter address")
	cmd.Flags().String(flags.FlagGasPrices, "", "gas prices for a transaction")
	cmd.Flags().String(flagLogLevel, "info", "set the logger level")
	cmd.Flags().Uint64(flagMaxMessages, 10, "The maximum number of messages in a transaction")
	cmd.Flags().String(flagBroadcastTimeout, "5m", "The time that cylinder will wait for tx commit")
	cmd.Flags().String(flagRPCPollInterval, "1s", "The duration of rpc poll interval")
	cmd.Flags().Uint64(flagMaxTry, 5, "The maximum number of tries to submit a transaction")
	cmd.Flags().Uint64(flagMinDE, 5, "The minimum number of DE")
	cmd.Flags().Float64(flagGasAdjustStart, 1.6, "The start value of gas adjustment")
	cmd.Flags().Float64(flagGasAdjustStep, 0.2, "The increment step of gad adjustment")
	cmd.Flags().BytesHex(flagRandomSecret, nil, "The secret value that is used for random D,E")
	cmd.Flags().Duration(flagCheckingDEInterval, 5*time.Minute, "The interval of checking DE")
	cmd.Flags().String(flagMetricsListenAddr, "", "address to use for metrics server.")

	flagNames := []string{
		flags.FlagChainID, flags.FlagNode, flagGranter, flags.FlagGasPrices, flagLogLevel,
		flagMaxMessages, flagBroadcastTimeout, flagRPCPollInterval, flagMaxTry, flagMinDE,
		flagGasAdjustStart, flagGasAdjustStep, flagRandomSecret, flagCheckingDEInterval, flagMetricsListenAddr,
	}

	for _, flagName := range flagNames {
		if err := viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName)); err != nil {
			panic(err)
		}
	}

	return cmd
}
