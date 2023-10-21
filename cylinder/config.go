package cylinder

import (
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// Config data structure for Cylinder process.
type Config struct {
	ChainID          string        `mapstructure:"chain-id"`          // ChainID of the target chain
	NodeURI          string        `mapstructure:"node"`              // Remote RPC URI of BandChain node to connect to
	Granter          string        `mapstructure:"granter"`           // The granter address
	GasPrices        string        `mapstructure:"gas-prices"`        // Gas prices of the transaction
	LogLevel         string        `mapstructure:"log-level"`         // Log level of the logger
	MaxMessages      uint64        `mapstructure:"max-messages"`      // The maximum number of messages in a transaction
	BroadcastTimeout time.Duration `mapstructure:"broadcast-timeout"` // The time that cylinder will wait for tx commit
	RPCPollInterval  time.Duration `mapstructure:"rpc-poll-interval"` // The duration of rpc poll interval
	MaxTry           uint64        `mapstructure:"max-try"`           // The maximum number of tries to submit a report transaction
	MinDE            uint64        `mapstructure:"min-de"`            // The minimum number of DE
	GasAdjustStart   float64       `mapstructure:"gas-adjust-start"`  // The start value of gas adjustment
	GasAdjustStep    float64       `mapstructure:"gas-adjust-step"`   // The increment step of gad adjustment
	RandomSecret     tss.Scalar    `mapstructure:"random-secret"`     // The secret value that is used for random D,E
	ActivePeriod     time.Duration `mapstructure:"active-period"`     // The time period that cylinder will send active status to chain
}
