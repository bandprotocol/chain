package emitter

import (
	"math"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ConvertToGas(owasm uint64) uint64 {
	// TODO: Using `gasConversionFactor` from oracle module
	return uint64(math.Ceil(float64(owasm) / float64(20_000_000)))
}

func MustParseValAddress(addr string) sdk.ValAddress {
	val, err := sdk.ValAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return val
}

func splitTxEvents(msgSize int, events []abci.Event) [][]abci.Event {
	eventGroups := make([][]abci.Event, msgSize)
	for _, event := range events {
		n := len(event.Attributes)
		attrType := event.Attributes[n-1]
		if attrType.Key != "msg_index" {
			panic("The last attribute of tx event should be msg_index")
		}
		idx, err := strconv.ParseInt(attrType.Value, 10, 0)
		if err != nil {
			panic(err)
		}
		eventGroups[idx] = append(eventGroups[idx], event)
	}
	return eventGroups
}
