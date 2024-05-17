package cmd

import (
	"errors"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/grogu/feed"
	"github.com/bandprotocol/chain/v2/grogu/priceservice"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	flagValidator                   = "validator"
	flagLogLevel                    = "log-level"
	flagPriceService                = "price-service"
	flagBroadcastTimeout            = "broadcast-timeout"
	flagRPCPollInterval             = "rpc-poll-interval"
	flagMaxTry                      = "max-try"
	flagDistributionStartPercentage = "distribution-start"
	flagDistributionPercentageRange = "distribution-range"
)

func runImpl(c *grogucontext.Context) error {
	c.Logger.Info(":rocket: Starting WebSocket subscriber")
	err := c.Client.Start()
	if err != nil {
		return err
	}
	defer c.Client.Stop() //nolint:errcheck

	for i := int64(0); i < int64(len(c.Keys)); i++ {
		c.FreeKeys <- i
	}

	c.Logger.Info(":rocket: Starting Prices submitter")
	go feed.StartSubmitPrices(c)

	c.Logger.Info(":rocket: Starting Prices querier")
	go feed.StartQuerySignalIDs(c)

	c.Logger.Info(":rocket: Starting Feed checker")
	feed.StartCheckFeeds(c)

	return nil
}

func RunCmd(c *grogucontext.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the grogu process",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := c.Config
			if config.ChainID == "" {
				return errors.New("chain ID must not be empty")
			}
			keys, err := c.Keyring.List()
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				return errors.New("no key available")
			}
			c.Keys = keys
			c.Validator, err = sdk.ValAddressFromBech32(config.Validator)
			if err != nil {
				return err
			}
			err = sdk.VerifyAddressFormat(c.Validator)
			if err != nil {
				return err
			}

			c.GasPrices = config.GasPrices

			allowLevel, err := log.AllowLevel(config.LogLevel)
			if err != nil {
				return err
			}
			c.Logger = logger.New(allowLevel)
			c.PriceService, err = priceservice.PriceServiceFromUrl(config.PriceService)
			if err != nil {
				return err
			}
			c.Logger.Info(":star: Creating HTTP client with node URI: %s", config.NodeURI)
			c.Client, err = httpclient.New(config.NodeURI, "/websocket")
			if err != nil {
				return err
			}
			cdc := band.MakeEncodingConfig().Marshaler
			clientCtx := client.Context{
				Client:            c.Client,
				Codec:             cdc,
				TxConfig:          band.MakeEncodingConfig().TxConfig,
				BroadcastMode:     flags.BroadcastSync,
				InterfaceRegistry: band.MakeEncodingConfig().InterfaceRegistry,
			}
			c.QueryClient = types.NewQueryClient(clientCtx)
			c.BroadcastTimeout, err = time.ParseDuration(config.BroadcastTimeout)
			if err != nil {
				return err
			}
			c.MaxTry = config.MaxTry
			c.RPCPollInterval, err = time.ParseDuration(config.RPCPollInterval)
			if err != nil {
				return err
			}
			c.FreeKeys = make(chan int64, len(keys))
			c.InProgressSignalIDs = &sync.Map{}
			c.PendingSignalIDs = make(chan map[string]time.Time, 100)
			c.PendingPrices = make(chan []types.SubmitPrice, 30)
			return runImpl(c)
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "chain ID of BandChain network")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "RPC url to BandChain node")
	cmd.Flags().String(flagValidator, "", "validator address")
	cmd.Flags().String(flagPriceService, "", "price-service name and url for getting prices")
	cmd.Flags().String(flags.FlagGasPrices, "", "gas prices for a transaction")
	cmd.Flags().String(flagLogLevel, "info", "set the logger level")
	cmd.Flags().String(flagBroadcastTimeout, "5m", "The time that Grogu will wait for tx commit")
	cmd.Flags().String(flagRPCPollInterval, "1s", "The duration of rpc poll interval")
	cmd.Flags().Uint64(flagMaxTry, 5, "The maximum number of tries to submit a transaction")
	cmd.Flags().
		Uint64(flagDistributionStartPercentage, 50, "The starting percentage of the distribution range of price sending")
	cmd.Flags().
		Uint64(flagDistributionPercentageRange, 30, "The ending percentage of the distribution range of price sending")
	_ = viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))
	_ = viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	_ = viper.BindPFlag(flagValidator, cmd.Flags().Lookup(flagValidator))
	_ = viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))
	_ = viper.BindPFlag(flagLogLevel, cmd.Flags().Lookup(flagLogLevel))
	_ = viper.BindPFlag(flagPriceService, cmd.Flags().Lookup(flagPriceService))
	_ = viper.BindPFlag(flagBroadcastTimeout, cmd.Flags().Lookup(flagBroadcastTimeout))
	_ = viper.BindPFlag(flagRPCPollInterval, cmd.Flags().Lookup(flagRPCPollInterval))
	_ = viper.BindPFlag(flagMaxTry, cmd.Flags().Lookup(flagMaxTry))
	_ = viper.BindPFlag(flagDistributionStartPercentage, cmd.Flags().Lookup(flagDistributionStartPercentage))
	_ = viper.BindPFlag(flagDistributionPercentageRange, cmd.Flags().Lookup(flagDistributionPercentageRange))

	return cmd
}
