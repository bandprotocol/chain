package grogu

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bandprotocol/chain/v2/grogu/executor"
	"github.com/bandprotocol/chain/v2/pkg/filecache"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	TxQuery = "tm.event = 'Tx' AND request.id EXISTS"
	// EventChannelCapacity is a buffer size of channel between node and this program
	EventChannelCapacity = 2000
)

// InProgressSymbols represents a data structure to store symbols in progress.
type InProgressSymbols struct {
	mu      sync.Mutex
	symbols map[string]time.Time
}

// MarkInProgress adds a symbol to the in-progress list with the current time.
func (ips *InProgressSymbols) MarkInProgress(symbol string) {
	ips.mu.Lock()
	defer ips.mu.Unlock()
	ips.symbols[symbol] = time.Now()
}

// MarkCompleted removes a symbol from the in-progress list.
func (ips *InProgressSymbols) MarkCompleted(symbol string) {
	ips.mu.Lock()
	defer ips.mu.Unlock()
	delete(ips.symbols, symbol)
}

// GetInProgressSymbols returns a list of symbols currently in progress.
func (ips *InProgressSymbols) GetInProgressSymbols() []string {
	ips.mu.Lock()
	defer ips.mu.Unlock()
	var inProgress []string
	for symbol := range ips.symbols {
		inProgress = append(inProgress, symbol)
	}
	return inProgress
}

// IsSymbolInProgress checks if a symbol is currently in progress.
func (ips *InProgressSymbols) IsSymbolInProgress(symbol string) bool {
	ips.mu.Lock()
	defer ips.mu.Unlock()
	_, inProgress := ips.symbols[symbol]
	return inProgress
}

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

	bz := cdc.MustMarshal(&types.QuerySymbolsRequest{})
	resBz, err := c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/Symbols", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get symbols with error: %s", c, err.Error())
	}
	symbols := types.QuerySymbolsResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &symbols)

	var symbolList []string
	for _, symbol := range symbols.Symbols {
		symbolList = append(symbolList, symbol.Symbol)
	}

	// symbolStr := strings.Join(symbolList, ",")

	// mockParams := map[string]string{
	// 	"symbols": symbolStr,
	// }

	for {
		l.Info("for loop")
		// keyIndex := <-c.freeKeys
		// l.Info("get keyIndex")

		// prices, err := c.executor.Exec(mockParams)
		// if err != nil {
		// 	fmt.Println("exec err", err)
		// } else {
		// 	fmt.Println("exec res", prices)
		// }
		// go SubmitPrices(c, l, keyIndex, prices)
		checkSymbols(c, l)
		time.Sleep(time.Second)
	}
}

func checkSymbols(c *Context, l *Logger) {
	bz := cdc.MustMarshal(&types.QuerySymbolsRequest{})
	resBz, err := c.client.ABCIQuery(context.Background(), "/feeds.v1beta1.Query/Symbols", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get symbols with error: %s", c, err.Error())
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
	}
	validatorPricesResponse := types.QueryValidatorPricesResponse{}
	cdc.MustUnmarshal(resBz.Response.Value, &validatorPricesResponse)
	validatorPrices := validatorPricesResponse.ValidatorPrices
	now := time.Now()

	for _, symbol := range symbols {
		if !c.inProgressSymbols.IsSymbolInProgress(symbol.GetSymbol()) {
			validatorPrice := findValidatorPrice(symbol.GetSymbol(), validatorPrices)
			if validatorPrice == nil ||
				time.Unix(validatorPrice.GetTimestamp(), 0).
					Add(time.Duration(symbol.MinInterval)*time.Second).
					Before(now) {
				symbolList = append(symbolList, symbol.Symbol)
				c.inProgressSymbols.MarkInProgress(symbol.GetSymbol())
			}
		}
	}

	fmt.Println("symbols list", symbolList)
}

func query_symbols(symbolList []string) {

}

func findValidatorPrice(symbol string, validatorPrices []types.PriceValidator) *types.PriceValidator {
	for _, validatorPrice := range validatorPrices {
		if validatorPrice.Symbol == symbol {
			return &validatorPrice
		}
	}
	return nil
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
			c.inProgressSymbols = &InProgressSymbols{
				symbols: make(map[string]time.Time),
			}
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
