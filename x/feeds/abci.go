package feeds

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/keeper"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// HandleEndBlock is a handler function for the EndBlock ABCI request.
func HandleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	symbols := k.GetSupportedSymbolsByPower(ctx)
	for _, symbol := range symbols {
		price, err := k.CalculatePrice(ctx, symbol)
		if err != nil {
			// TODO: handle error
			continue
		}

		k.SetPrice(ctx, price)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUpdatePrice,
				sdk.NewAttribute(types.AttributeKeySymbol, price.Symbol),
				sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", price.Price)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", price.Timestamp)),
			),
		)
	}
}
