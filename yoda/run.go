package yoda

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/filecache"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
	"github.com/bandprotocol/chain/v3/yoda/executor"
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
	defer c.client.Stop() //nolint:errcheck

	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()

	l.Info(":ear: Subscribing to events with query: %s...", TxQuery)
	eventChan, err := c.client.Subscribe(ctx, "", TxQuery, EventChannelCapacity)
	if err != nil {
		return err
	}

	if c.metricsEnabled {
		l.Info(":eyes: Starting Prometheus listener")
		go metricsListen(cfg.MetricsListenAddr, c)
	}

	availableKeys := make([]bool, len(c.keys))
	waitingMsgs := make([][]ReportMsgWithKey, len(c.keys))
	for i := range availableKeys {
		availableKeys[i] = true
		waitingMsgs[i] = []ReportMsgWithKey{}
	}

	bz := c.encodingConfig.Codec.MustMarshal(&types.QueryPendingRequestsRequest{
		ValidatorAddress: c.validator.String(),
	})
	resBz, err := c.client.ABCIQuery(context.Background(), "/band.oracle.v1.Query/PendingRequests", bz)
	if err != nil {
		l.Error(":exploding_head: Failed to get pending requests with error: %s", c, err.Error())
	}
	pendingRequests := types.QueryPendingRequestsResponse{}
	c.encodingConfig.Codec.MustUnmarshal(resBz.Response.Value, &pendingRequests)

	l.Info(":mag: Found %d pending requests", len(pendingRequests.RequestIDs))
	for _, id := range pendingRequests.RequestIDs {
		c.pendingRequests[types.RequestID(id)] = true
		go handleRequest(c, l, types.RequestID(id))
	}

	for {
		select {
		case ev := <-eventChan:
			go handleTransaction(c, l, ev.Data.(cmttypes.EventDataTx).TxResult)
		case keyIndex := <-c.freeKeys:
			if len(waitingMsgs[keyIndex]) != 0 {
				if uint64(len(waitingMsgs[keyIndex])) > c.maxReport {
					go SubmitReport(c, l, keyIndex, waitingMsgs[keyIndex][:c.maxReport])
					waitingMsgs[keyIndex] = waitingMsgs[keyIndex][c.maxReport:]
				} else {
					go SubmitReport(c, l, keyIndex, waitingMsgs[keyIndex])
					waitingMsgs[keyIndex] = []ReportMsgWithKey{}
				}
			} else {
				availableKeys[keyIndex] = true
			}
		case pm := <-c.pendingMsgs:
			c.updatePendingGauge(1)
			if availableKeys[pm.keyIndex] {
				availableKeys[pm.keyIndex] = false
				go SubmitReport(c, l, pm.keyIndex, []ReportMsgWithKey{pm})
			} else {
				waitingMsgs[pm.keyIndex] = append(waitingMsgs[pm.keyIndex], pm)
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

			allowLevel, err := log.ParseLogLevel(cfg.LogLevel)
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
			c.pendingMsgs = make(chan ReportMsgWithKey)
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
	cmd.Flags().String(flagBroadcastTimeout, "5m", "The time that Yoda will wait for tx commit")
	cmd.Flags().String(flagRPCPollInterval, "1s", "The duration of rpc poll interval")
	cmd.Flags().Uint64(flagMaxTry, 5, "The maximum number of tries to submit a report transaction")
	cmd.Flags().Uint64(flagMaxReport, 10, "The maximum number of reports in one transaction")
	_ = viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID))
	_ = viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	_ = viper.BindPFlag(flagValidator, cmd.Flags().Lookup(flagValidator))
	_ = viper.BindPFlag(flags.FlagGasPrices, cmd.Flags().Lookup(flags.FlagGasPrices))
	_ = viper.BindPFlag(flagLogLevel, cmd.Flags().Lookup(flagLogLevel))
	_ = viper.BindPFlag(flagExecutor, cmd.Flags().Lookup(flagExecutor))
	_ = viper.BindPFlag(flagBroadcastTimeout, cmd.Flags().Lookup(flagBroadcastTimeout))
	_ = viper.BindPFlag(flagRPCPollInterval, cmd.Flags().Lookup(flagRPCPollInterval))
	_ = viper.BindPFlag(flagMaxTry, cmd.Flags().Lookup(flagMaxTry))
	_ = viper.BindPFlag(flagMaxReport, cmd.Flags().Lookup(flagMaxReport))

	return cmd
}
