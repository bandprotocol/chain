package grogu

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bandprotocol/chain/v2/grogu/executor"
	"github.com/bandprotocol/chain/v2/pkg/filecache"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	TxQuery = "tm.event = 'Tx' AND request.id EXISTS"
	// EventChannelCapacity is a buffer size of channel between node and this program
	EventChannelCapacity = 2000
)

func runImpl(c *Context, l *Logger) error {
	l.Info(":rocket: Starting WebSocket subscriber")
	err := c.client.Start()
	if err != nil {
		return err
	}
	defer c.client.Stop()

	// ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cxl()

	l.Info(":ear: Subscribing to events with query: %s...", TxQuery)
	// eventChan, err := c.client.Subscribe(ctx, "", TxQuery, EventChannelCapacity)
	// if err != nil {
	// 	return err
	// }

	if c.metricsEnabled {
		l.Info(":eyes: Starting Prometheus listener")
		go metricsListen(cfg.MetricsListenAddr, c)
	}

	availiableKeys := make([]bool, len(c.keys))
	waitingMsgs := make([][]ReportMsgWithKey, len(c.keys))
	for i := range availiableKeys {
		availiableKeys[i] = true
		waitingMsgs[i] = []ReportMsgWithKey{}
	}
	l.Info("len freekeys", len(c.freeKeys))
	for i := int64(0); i < int64(len(c.keys)); i++ {
		l.Info("put key")
		c.freeKeys <- i
	}
	l.Info("finished put key")

	bz := cdc.MustMarshal(&feedstypes.QuerySymbolsRequest{})
	resBz, err := c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/Symbols", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get symbols with error: %s", c, err.Error())
	}
	symbols := feedstypes.QuerySymbolsResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &symbols)

	var symbolList []string
	for _, symbol := range symbols.Symbols {
		symbolList = append(symbolList, symbol.Symbol)
	}

	symbolStr := strings.Join(symbolList, ",")

	mockParams := map[string]string{
		"symbols": symbolStr,
	}

	for {
		l.Info("for loop")
		keyIndex := <-c.freeKeys
		l.Info("get keyIndex")

		prices, err := c.executor.Exec(mockParams)
		if err != nil {
			fmt.Println("exec err", err)
		} else {
			fmt.Println("exec res", prices)
		}
		go SubmitPrices(c, l, keyIndex, prices)
		checkSymbol(c, l)
		time.Sleep(time.Second)
	}
}

func checkSymbol(c *Context, l *Logger) {
	bz := cdc.MustMarshal(&feedstypes.QuerySymbolsRequest{})
	resBz, err := c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/Symbols", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get symbols with error: %s", c, err.Error())
	}
	symbolsResponse := feedstypes.QuerySymbolsResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &symbolsResponse)

	symbols := symbolsResponse.Symbols

	now := time.Now()
	var symbolList []string

	for _, symbol := range symbols {
		if time.Unix(symbol.Timestamp, 0).Add(time.Duration(symbol.Interval) * time.Second).Before(now) {
			symbolList = append(symbolList, symbol.Symbol)
		}
	}

	c.pendingSymbols <- symbolList
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
			c.executor, err = executor.NewExecutor(cfg.Executor)
			if err != nil {
				return err
			}
			l.Info(":star: Creating HTTP client with node URI: %s", cfg.NodeURI)
			c.client, err = httpclient.New(cfg.NodeURI, "/websocket")
			if err != nil {
				return err
			}
			c.fileCache = filecache.New(filepath.Join(c.home, "files"))
			c.broadcastTimeout, err = time.ParseDuration(cfg.BroadcastTimeout)
			if err != nil {
				return err
			}
			c.maxTry = cfg.MaxTry
			c.maxReport = cfg.MaxReport
			c.rpcPollInterval, err = time.ParseDuration(cfg.RPCPollInterval)
			if err != nil {
				return err
			}
			c.pendingSymbols = make(chan []string)
			c.freeKeys = make(chan int64, len(keys))
			c.keyRoundRobinIndex = -1
			c.pendingRequests = make(map[types.RequestID]bool)
			c.metricsEnabled = cfg.MetricsListenAddr != ""
			return runImpl(c, l)
		},
	}
	cmd.Flags().String(flags.FlagChainID, "", "chain ID of BandChain network")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "RPC url to BandChain node")
	cmd.Flags().String(flagValidator, "", "validator address")
	cmd.Flags().String(flagExecutor, "", "executor name and url for executing the data source script")
	cmd.Flags().String(flags.FlagGasPrices, "", "gas prices for report transaction")
	cmd.Flags().String(flagLogLevel, "info", "set the logger level")
	cmd.Flags().String(flagBroadcastTimeout, "5m", "The time that Grogu will wait for tx commit")
	cmd.Flags().String(flagRPCPollInterval, "1s", "The duration of rpc poll interval")
	cmd.Flags().Uint64(flagMaxTry, 5, "The maximum number of tries to submit a report transaction")
	cmd.Flags().Uint64(flagMaxReport, 10, "The maximum number of reports in one transaction")
	viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	viper.BindPFlag(flagValidator, cmd.Flags().Lookup(flagValidator))
	viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))
	viper.BindPFlag(flagLogLevel, cmd.Flags().Lookup(flagLogLevel))
	viper.BindPFlag(flagExecutor, cmd.Flags().Lookup(flagExecutor))
	viper.BindPFlag(flagBroadcastTimeout, cmd.Flags().Lookup(flagBroadcastTimeout))
	viper.BindPFlag(flagRPCPollInterval, cmd.Flags().Lookup(flagRPCPollInterval))
	viper.BindPFlag(flagMaxTry, cmd.Flags().Lookup(flagMaxTry))
	viper.BindPFlag(flagMaxReport, cmd.Flags().Lookup(flagMaxReport))
	return cmd
}
