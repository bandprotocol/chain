package emitter

import (
	"math"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
)

func ConvertToGas(owasm uint64) uint64 {
	return uint64(math.Ceil(float64(owasm) / float64(oraclekeeper.GasConversionFactor)))
}

func MustParseValAddress(addr string) sdk.ValAddress {
	val, err := sdk.ValAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return val
}

func splitTxEvents(msgSize int, events []abci.Event) ([]abci.Event, [][]abci.Event) {
	var txEvents []abci.Event
	eventGroups := make([][]abci.Event, msgSize)
	for _, event := range events {
		n := len(event.Attributes)
		attrType := event.Attributes[n-1]
		if attrType.Key != "msg_index" {
			txEvents = append(txEvents, event)
		} else {
			idx, err := strconv.ParseInt(attrType.Value, 10, 0)
			if err != nil {
				panic(err)
			}
			eventGroups[idx] = append(eventGroups[idx], event)
		}
	}
	return txEvents, eventGroups
}
