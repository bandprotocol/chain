package keeper

import (
	"github.com/bandprotocol/chain/v2/x/tssmember/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) types.QueryServer {
	return queryServer{k: k}
}
