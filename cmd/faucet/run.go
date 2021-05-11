package main

import (
	"errors"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
)

func runLimiter() {
	uptimeTicker := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-uptimeTicker.C:
			toRemove := make([]string, 0, 10)
			for k, v := range limit.status.container {
				denomsUnpend := 0
				for _, vw := range v.LastWithdrawals {
					if time.Now().Sub(vw) > cfg.Period {
						denomsUnpend++
					}
				}
				if denomsUnpend == len(v.LastWithdrawals) {
					toRemove = append(toRemove, k)
				}
			}
			for _, k := range toRemove {
				limit.status.Remove(k)
			}
		}
	}
}

func runCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the oracle process",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.ChainID == "" {
				return errors.New("chain ID must not be empty")
			}
			keys, err := keybase.List()
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				return errors.New("No key available")
			}
			c.keys = make(chan keyring.Info, len(keys))
			for _, key := range keys {
				c.keys <- key
			}
			c.gasPrices, err = sdk.ParseDecCoins(cfg.GasPrices)
			if err != nil {
				return err
			}
			c.client, err = httpclient.New(cfg.NodeURI, "/websocket")
			if err != nil {
				return err
			}
			r := gin.Default()
			r.Use(func(c *gin.Context) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				c.Writer.Header().Set("Access-Control-Allow-Methods", "POST")

				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(204)
					return
				}
			})

			c.coins, err = sdk.ParseCoinsNormalized(cfg.Coins)
			if err != nil {
				panic(err)
			}

			c.maxPerPeriodWithdrawal, err = sdk.ParseCoinsNormalized(cfg.MaxPerPeriodWithdrawal)
			if err != nil {
				panic(err)
			}

			limit = NewLimit(c, cfg)

			r.POST("/request", func(gc *gin.Context) {
				handleRequest(gc, c)
			})

			go runLimiter()

			return r.Run("0.0.0.0:" + cfg.Port)
		},
	}
	cmd.Flags().String(flags.FlagChainID, "", "chain ID of BandChain network")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "RPC url to BandChain node")
	cmd.Flags().String(flags.FlagGasPrices, "", "gas prices for report transaction")
	cmd.Flags().String(flagPort, "5005", "port of faucet service")
	cmd.Flags().String(flagCoins, "10loki", "coins to create")
	cmd.Flags().Duration(flagPeriod, 12*time.Hour, "period when can withdraw again")
	viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))
	viper.BindPFlag(flagPort, cmd.Flags().Lookup(flagPort))
	viper.BindPFlag(flagCoins, cmd.Flags().Lookup(flagCoins))
	viper.BindPFlag(flagPeriod, cmd.Flags().Lookup(flagPeriod))
	return cmd
}
