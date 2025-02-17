package context

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/bandprotocol/chain/v3/app/params"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

// Config holds the configuration settings for the application.
type Config struct {
	// Validator is the address used to submit prices.
	Validator string `mapstructure:"validator"`

	// NodeURIs are the BandChain RPC URIs to connect to.
	NodeURIs string `mapstructure:"nodes"`

	// ChainID is the target BandChain chain ID.
	ChainID string `mapstructure:"chain-id"`

	// BroadcastTimeout is the duration Grogu will wait for a transaction commit.
	BroadcastTimeout string `mapstructure:"broadcast-timeout"`

	// RPCPollInterval is the duration between RPC polls.
	RPCPollInterval string `mapstructure:"rpc-poll-interval"`

	// MaxTry is the maximum number of attempts to submit a transaction.
	MaxTry uint64 `mapstructure:"max-try"`

	// GasPrices is the gas price set for each transaction.
	GasPrices string `mapstructure:"gas-prices"`

	// DistributionStartPercentage defines the initial percentage for price distribution.
	DistributionStartPercentage uint64 `mapstructure:"distribution-start-pct"`

	// DistributionOffsetPercentage defines the range of the percentage for price distribution.
	DistributionOffsetPercentage uint64 `mapstructure:"distribution-offset-pct"`

	// Bothan is the URL for connecting to Bothan.
	Bothan string `mapstructure:"bothan"`

	// BothanTimeout is the timeout duration for Bothan requests.
	BothanTimeout string `mapstructure:"bothan-timeout"`

	// LogLevel is the level of logging for the logger.
	LogLevel string `mapstructure:"log-level"`

	// UpdaterQueryInterval is the interval for updater querying chain.
	UpdaterQueryInterval string `mapstructure:"updater-query-interval"`
}

// Context holds the runtime context for the application.
type Context struct {
	Config         Config
	Keyring        keyring.Keyring
	Logger         *logger.Logger
	Home           string
	EncodingConfig params.EncodingConfig
}
