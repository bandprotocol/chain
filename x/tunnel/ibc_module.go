package tunnel

import (
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
)

type IBCModule struct {
	keeper keeper.Keeper
}

func NewIBCModule(keeper keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: keeper,
	}
}

// func (im IBCModule) OnChanOpenInit(ctx sdk.Context,
// 	order channeltypes.Order,
// 	connectionHops []string,
// 	portID string,
// 	channelID string,
// 	channelCap *capabilitytypes.Capability,
// 	counterparty channeltypes.Counterparty,
// 	version string,
// ) (string, error) {

// }
