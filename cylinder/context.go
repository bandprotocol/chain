package cylinder

import (
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/logger"
)

// Context holds the context information for the Cylinder process.
type Context struct {
	Config  *Config
	Keyring keyring.Keyring
	Home    string

	Logger *logger.Logger

	ErrCh chan error
	MsgCh chan sdk.Msg

	Store *store.Store
}

// NewContext creates a new instance of the Context.
func NewContext(cfg *Config, kr keyring.Keyring, home string) (*Context, error) {
	// Initialize the context
	ctx := &Context{
		Config:  cfg,
		Keyring: kr,
		Home:    home,
	}

	// Create the logger
	allowLevel, err := log.AllowLevel(ctx.Config.LogLevel)
	if err != nil {
		return nil, err
	}
	ctx.Logger = logger.NewLogger(allowLevel)

	// Create the error and message channels
	ctx.ErrCh = make(chan error, 1)
	ctx.MsgCh = make(chan sdk.Msg, 1000)

	// Create the store
	dataDir := filepath.Join(ctx.Home, "data")
	db, err := dbm.NewDB("cylinder", dbm.GoLevelDBBackend, dataDir)
	if err != nil {
		return nil, err
	}
	ctx.Store = store.NewStore(db)

	return ctx, nil
}
