package context

import (
	"fmt"
	"path/filepath"
	"time"

	dbm "github.com/cometbft/cometbft-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder/store"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/pkg/tss"
)

// Config data structure for Cylinder process.
type Config struct {
	ChainID            string        `mapstructure:"chain-id"`             // ChainID of the target chain
	NodeURI            string        `mapstructure:"node"`                 // Remote RPC URI of BandChain node to connect to
	Granter            string        `mapstructure:"granter"`              // The granter address
	GasPrices          string        `mapstructure:"gas-prices"`           // Gas prices of the transaction
	LogLevel           string        `mapstructure:"log-level"`            // Log level of the logger
	MaxMessages        uint64        `mapstructure:"max-messages"`         // The maximum number of messages in a transaction
	BroadcastTimeout   time.Duration `mapstructure:"broadcast-timeout"`    // The time that cylinder will wait for tx commit
	RPCPollInterval    time.Duration `mapstructure:"rpc-poll-interval"`    // The duration of rpc poll interval
	MaxTry             uint64        `mapstructure:"max-try"`              // The maximum number of tries to submit a report transaction
	MinDE              uint64        `mapstructure:"min-de"`               // The minimum number of DE
	GasAdjustStart     float64       `mapstructure:"gas-adjust-start"`     // The start value of gas adjustment
	GasAdjustStep      float64       `mapstructure:"gas-adjust-step"`      // The increment step of gad adjustment
	RandomSecret       tss.Scalar    `mapstructure:"random-secret"`        // The secret value that is used for random D,E
	CheckingDEInterval time.Duration `mapstructure:"checking-de-interval"` // The interval for updating DE
}

// Context holds the context information for the Cylinder process.
type Context struct {
	Config            *Config
	Keyring           keyring.Keyring
	Home              string
	Cdc               codec.Codec             // Codec for serialization.
	TxConfig          client.TxConfig         // Transaction configuration.
	InterfaceRegistry types.InterfaceRegistry // Interface registry for protobuf types.

	Logger *logger.Logger

	ErrCh chan error
	MsgCh chan sdk.Msg

	DataDir string
	Store   *store.Store
}

// NewContext creates a new instance of the Context.
func NewContext(
	cfg *Config,
	kr keyring.Keyring,
	home string,
	cdc codec.Codec,
	txConfig client.TxConfig,
	interfaceRegistry types.InterfaceRegistry,
) (*Context, error) {
	// Create the store
	dataDir := filepath.Join(home, "data")

	// Initialize the context
	return &Context{
		Config:            cfg,
		Keyring:           kr,
		Home:              home,
		Cdc:               cdc,
		TxConfig:          txConfig,
		InterfaceRegistry: interfaceRegistry,
		ErrCh:             make(chan error, 1),
		MsgCh:             make(chan sdk.Msg, 1000),
		DataDir:           dataDir,
	}, nil
}

func (ctx *Context) InitLog() error {
	allowLevel, err := log.ParseLogLevel(ctx.Config.LogLevel)
	if err != nil {
		return err
	}

	ctx.Logger = logger.NewLogger(allowLevel)
	return nil
}

// WithGoLevelDb initializes the database of the context with GoLevelDB.
func (ctx *Context) WithGoLevelDb() (*Context, error) {
	db, err := dbm.NewDB("cylinder", dbm.GoLevelDBBackend, ctx.DataDir)
	if err != nil {
		return nil, err
	}

	return ctx.WithDB(db)
}

// WithDB sets the DB for the context.
func (ctx *Context) WithDB(db dbm.DB) (*Context, error) {
	if ctx.Store != nil {
		if err := ctx.Store.DB.Close(); err != nil {
			return nil, fmt.Errorf("failed to close the existing DB: %w", err)
		}
	}

	ctx.Store = store.NewStore(db)
	return ctx, nil
}
