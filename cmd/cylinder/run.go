package main

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/workers/round1"
	"github.com/bandprotocol/chain/v2/cylinder/workers/round2"
	"github.com/bandprotocol/chain/v2/cylinder/workers/sender"
)

const (
	flagGranter          = "granter"
	flagLogLevel         = "log-level"
	flagBroadcastTimeout = "broadcast-timeout"
	flagRPCPollInterval  = "rpc-poll-interval"
	flagMaxTry           = "max-try"
)

// runCmd returns a Cobra command to run the cylinder process.
func runCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the cylinder process",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cylinder.NewContext(ctx.config, ctx.keyring, ctx.home)
			if err != nil {
				return err
			}

			round1, err := round1.New(c)
			if err != nil {
				return err
			}

			round2, err := round2.New(c)
			if err != nil {
				return err
			}

			sender, err := sender.New(c)
			if err != nil {
				return err
			}

			workers := cylinder.Workers{round1, round2, sender}

			return cylinder.Run(c, workers)
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "chain ID of BandChain network")
	viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))

	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "RPC url to BandChain node")
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))

	cmd.Flags().String(flagGranter, "", "granter address")
	viper.BindPFlag(flagGranter, cmd.Flags().Lookup(flagGranter))

	cmd.Flags().String(flags.FlagGasPrices, "", "gas prices for a transaction")
	viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))

	cmd.Flags().String(flagLogLevel, "info", "set the logger level")
	viper.BindPFlag(flagLogLevel, cmd.Flags().Lookup(flagLogLevel))

	cmd.Flags().String(flagBroadcastTimeout, "5m", "The time that cylinder will wait for tx commit")
	viper.BindPFlag(flagBroadcastTimeout, cmd.Flags().Lookup(flagBroadcastTimeout))

	cmd.Flags().String(flagRPCPollInterval, "1s", "The duration of rpc poll interval")
	viper.BindPFlag(flagRPCPollInterval, cmd.Flags().Lookup(flagRPCPollInterval))

	cmd.Flags().Uint64(flagMaxTry, 5, "The maximum number of tries to submit a transaction")
	viper.BindPFlag(flagMaxTry, cmd.Flags().Lookup(flagMaxTry))

	return cmd
}
