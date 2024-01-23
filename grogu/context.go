package grogu

import (
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/grogu/executor"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

type Context struct {
	client           rpcclient.Client
	validator        sdk.ValAddress
	gasPrices        string
	keys             []*keyring.Record
	executor         executor.Executor
	broadcastTimeout time.Duration
	maxTry           uint64
	rpcPollInterval  time.Duration

	pendingPrices     chan []types.SubmitPrice
	inProgressSymbols *InProgressSymbols
	freeKeys          chan int64

	home string
}
