package context

import (
	"time"

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
	BroadcastTimeout time.Duration `mapstructure:"broadcast-timeout"`

	// RPCPollInterval is the duration between RPC polls.
	RPCPollInterval time.Duration `mapstructure:"rpc-poll-interval"`

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
	BothanTimeout time.Duration `mapstructure:"bothan-timeout"`

	// LogLevel is the level of logging for the logger.
	LogLevel string `mapstructure:"log-level"`

	// UpdaterQueryInterval is the interval for updater querying chain.
	UpdaterQueryInterval time.Duration `mapstructure:"updater-query-interval"`

	// MetricsListenAddr is an address to use for metrics server
	MetricsListenAddr string `mapstructure:"metrics-listen-addr"`
}

// Context holds the runtime context for the application.
type Context struct {
	Config         Config
	Keyring        keyring.Keyring
	Logger         *logger.Logger
	Home           string
	EncodingConfig params.EncodingConfig
}

func New(
	cfg Config,
	kr keyring.Keyring,
	logger *logger.Logger,
	home string,
	encodingConfig params.EncodingConfig,
) *Context {
	return &Context{
		Config:         cfg,
		Keyring:        kr,
		Logger:         logger,
		Home:           home,
		EncodingConfig: encodingConfig,
	}
}
