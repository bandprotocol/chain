package feed

import (
	"github.com/bandprotocol/chain/v2/x/feed/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HandleEndBlock is a handler function for the BeginBlock ABCI request.
func HandleEndBlock(ctx sdk.Context, k keeper.Keeper) {
}
