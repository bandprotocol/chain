package grogu

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bandprotocol/chain/v2/grogu/priceservice"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func runImpl(c *Context, l *Logger) error {
	l.Info(":rocket: Starting WebSocket subscriber")
	err := c.client.Start()
	if err != nil {
		return err
	}
	defer c.client.Stop()

	for i := int64(0); i < int64(len(c.keys)); i++ {
		c.freeKeys <- i
	}

	l.Info(":rocket: Starting Prices submitter")
	go startSubmitPrices(c, l)
	go startQuerySymbols(c, l)

	l.Info(":rocket: Starting Symbol checker")
	for {
		checkSymbols(c, l)
		time.Sleep(time.Second)
	}
}

func startSubmitPrices(c *Context, l *Logger) {
	for {
		SubmitPrices(c, l)
	}
}

func startQuerySymbols(c *Context, l *Logger) {
	for {
		querySymbols(c, l)
	}
}

func checkSymbols(c *Context, l *Logger) {
	bz := cdc.MustMarshal(&types.QuerySymbolsRequest{})
	resBz, err := c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/Symbols", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get symbols with error: %s", c, err.Error())
		return
	}

	symbolsResponse := types.QuerySymbolsResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &symbolsResponse)
	symbols := symbolsResponse.Symbols

	var symbolList []string

	bz = cdc.MustMarshal(&types.QueryValidatorPricesRequest{
		Validator: c.validator.String(),
	})
	resBz, err = c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/ValidatorPrices", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get validator prices with error: %s", c, err.Error())
		return
	}
	validatorPricesResponse := types.QueryValidatorPricesResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &validatorPricesResponse)
	validatorPrices := validatorPricesResponse.ValidatorPrices
	symbolTimestampMap := ConvertToSymbolTimestampMap(validatorPrices)

	now := time.Now()

	for _, symbol := range symbols {
		if _, inProgress := c.inProgressSymbols.Load(symbol.GetSymbol()); inProgress {
			continue
		}

		timestamp, ok := symbolTimestampMap[symbol.GetSymbol()]
		// add 2 to prevent too fast cases
		if !ok ||
			time.Unix(timestamp+2, 0).
				Add(time.Duration(symbol.MinInterval)*time.Second).
				Before(now) {
			symbolList = append(symbolList, symbol.Symbol)
			c.inProgressSymbols.Store(symbol.GetSymbol(), time.Now())
		}
	}
	if len(symbolList) != 0 {
		l.Info("found symbols to send: %v", symbolList)
		c.pendingSymbols <- symbolList
	}
}

// ConvertToSymbolTimestampMap converts an array of PriceValidator to a map of symbol to timestamp.
func ConvertToSymbolTimestampMap(data []types.PriceValidator) map[string]int64 {
	symbolTimestampMap := make(map[string]int64)

	for _, entry := range data {
		symbolTimestampMap[entry.Symbol] = entry.Timestamp
	}

	return symbolTimestampMap
}

// ConvertToSymbolPriceMap converts an array of SubmitPrice to a map of symbol to price.
func ConvertToSymbolPriceMap(data []types.SubmitPrice) map[string]uint64 {
	symbolPriceMap := make(map[string]uint64)

	for _, entry := range data {
		symbolPriceMap[entry.Symbol] = entry.Price
	}

	return symbolPriceMap
}

func querySymbols(c *Context, l *Logger) {
	symbols := <-c.pendingSymbols

GetAllSymbols:
	for {
		select {
		case nextSymbols := <-c.pendingSymbols:
			symbols = append(symbols, nextSymbols...)
		default:
			break GetAllSymbols
		}
	}

	symbolStr := strings.Join(symbols, ",")

	params := map[string]string{
		"symbols": symbolStr,
	}

	l.Info("Try to get prices for symbols: %s", symbolStr)
	prices, err := c.priceService.Query(params)
	if err != nil {
		l.Error(":exploding_head: Failed to get prices from price-service with error: %s", c, err.Error())
	}

	// delete symbol from in progress map if its price is not found
	symbolPriceMap := ConvertToSymbolPriceMap(prices)
	for _, symbol := range symbols {
		if _, found := symbolPriceMap[symbol]; !found {
			c.inProgressSymbols.Delete(symbol)
		}
	}

	l.Info("got prices for symbols: %s", symbolStr)
	if len(prices) == 0 {
		l.Error(":exploding_head: query symbol got no prices with symbols: %s", c, symbolStr)
		return
	}
	c.pendingPrices <- prices
}

func runCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run the grogu process",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.ChainID == "" {
				return errors.New("chain ID must not be empty")
			}
			keys, err := kb.List()
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				return errors.New("no key available")
			}
			c.keys = keys
			c.validator, err = sdk.ValAddressFromBech32(cfg.Validator)
			if err != nil {
				return err
			}
			err = sdk.VerifyAddressFormat(c.validator)
			if err != nil {
				return err
			}

			c.gasPrices = cfg.GasPrices

			allowLevel, err := log.AllowLevel(cfg.LogLevel)
			if err != nil {
				return err
			}
			l := NewLogger(allowLevel)
			c.priceService, err = priceservice.NewPriceService(cfg.PriceService)
			if err != nil {
				return err
			}
			l.Info(":star: Creating HTTP client with node URI: %s", cfg.NodeURI)
			c.client, err = httpclient.New(cfg.NodeURI, "/websocket")
			if err != nil {
				return err
			}
			c.broadcastTimeout, err = time.ParseDuration(cfg.BroadcastTimeout)
			if err != nil {
				return err
			}
			c.maxTry = cfg.MaxTry
			c.rpcPollInterval, err = time.ParseDuration(cfg.RPCPollInterval)
			if err != nil {
				return err
			}
			c.freeKeys = make(chan int64, len(keys))
			c.inProgressSymbols = &sync.Map{}
			c.pendingSymbols = make(chan []string, 100)
			c.pendingPrices = make(chan []types.SubmitPrice, 30)
			return runImpl(c, l)
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
	viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	viper.BindPFlag(flagValidator, cmd.Flags().Lookup(flagValidator))
	viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))
	viper.BindPFlag(flagLogLevel, cmd.Flags().Lookup(flagLogLevel))
	viper.BindPFlag(flagPriceService, cmd.Flags().Lookup(flagPriceService))
	viper.BindPFlag(flagBroadcastTimeout, cmd.Flags().Lookup(flagBroadcastTimeout))
	viper.BindPFlag(flagRPCPollInterval, cmd.Flags().Lookup(flagRPCPollInterval))
	viper.BindPFlag(flagMaxTry, cmd.Flags().Lookup(flagMaxTry))
	return cmd
}
