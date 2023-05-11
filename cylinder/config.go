package cylinder

import "time"

// Config data structure for cylinder daemon.
type Config struct {
	ChainID          string        `mapstructure:"chain-id"`          // ChainID of the target chain
	NodeURI          string        `mapstructure:"node"`              // Remote RPC URI of BandChain node to connect to
	Granter          string        `mapstructure:"granter"`           // The granter address that I'm responsible for
	GasPrices        string        `mapstructure:"gas-prices"`        // Gas prices of the transaction
	LogLevel         string        `mapstructure:"log-level"`         // Log level of the logger
	BroadcastTimeout time.Duration `mapstructure:"broadcast-timeout"` // The time that cylinder will wait for tx commit
	RPCPollInterval  time.Duration `mapstructure:"rpc-poll-interval"` // The duration of rpc poll interval
	MaxTry           uint64        `mapstructure:"max-try"`           // The maximum number of tries to submit a report transaction
}
