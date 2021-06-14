package main

import (
	"fmt"
	"github.com/GeoDB-Limited/odin-core/cmd/faucet/config"
	"github.com/GeoDB-Limited/odin-core/cmd/faucet/limiter"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagPort   = "port"
	flagCoins  = "coins"
	flagPeriod = "period"
)

const (
	DefaultChainID          = "odin"
	DefaultNodeURI          = "tcp://0.0.0.0:26657"
	DefaultGasPrices        = ""
	DefaultFaucetPort       = "5005"
	DefaultWithdrawalAmount = "10loki"
	DefaultFaucetPeriod     = 12 * time.Hour
)

// runCmd runs the faucet.
func runCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the oracle process",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := gin.Default()
			engine.Use(
				func(ginCtx *gin.Context) {
					ginCtx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
					ginCtx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
					ginCtx.Writer.Header().Set(
						"Access-Control-Allow-Headers",
						"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
					)
					ginCtx.Writer.Header().Set("Access-Control-Allow-Methods", "POST")

					if ginCtx.Request.Method == "OPTIONS" {
						ginCtx.AbortWithStatus(http.StatusNoContent)
						return
					}
				},
			)

			l := limiter.NewLimiter(cfg)
			go l.RunCleaner()

			engine.POST("/request", func(ginCtx *gin.Context) { l.HandleRequest(ginCtx) })
			return engine.Run(fmt.Sprintf(":%s", cfg.Port))
		},
	}

	cmd.Flags().String(flags.FlagChainID, DefaultChainID, "chain ID of BandChain network")
	if err := viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID)); err != nil {
		panic(err)
	}
	cmd.Flags().String(flags.FlagNode, DefaultNodeURI, "RPC url to Odin node")
	if err := viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode)); err != nil {
		panic(err)
	}
	cmd.Flags().String(flags.FlagGasPrices, DefaultGasPrices, "gas prices for report transaction")
	if err := viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices)); err != nil {
		panic(err)
	}
	cmd.Flags().String(flagPort, DefaultFaucetPort, "port of faucet service")
	if err := viper.BindPFlag(flagPort, cmd.Flags().Lookup(flagPort)); err != nil {
		panic(err)
	}
	cmd.Flags().String(flagCoins, DefaultWithdrawalAmount, "coins to create")
	if err := viper.BindPFlag(flagCoins, cmd.Flags().Lookup(flagCoins)); err != nil {
		panic(err)
	}
	cmd.Flags().Duration(flagPeriod, DefaultFaucetPeriod, "period when can withdraw again")
	if err := viper.BindPFlag(flagPeriod, cmd.Flags().Lookup(flagPeriod)); err != nil {
		panic(err)
	}

	return cmd
}
