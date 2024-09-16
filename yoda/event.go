package yoda

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

type rawRequest struct {
	dataSourceID   types.DataSourceID
	dataSourceHash string
	externalID     types.ExternalID
	calldata       string
}

// GetEventValues returns the list of all values in the given log with the given type and key.
func GetEventValues(events []abci.Event, evType string, evKey string) (res []string) {
	for _, ev := range events {
		if ev.Type != evType {
			continue
		}

		for _, attr := range ev.Attributes {
			if attr.Key == evKey {
				res = append(res, attr.Value)
			}
		}
	}
	return res
}

// GetEventValue checks and returns the exact value in the given log with the given type and key.
func GetEventValue(events []abci.Event, evType string, evKey string) (string, error) {
	values := GetEventValues(events, evType, evKey)
	if len(values) == 0 {
		return "", fmt.Errorf("cannot find event with type: %s, key: %s", evType, evKey)
	}
	if len(values) > 1 {
		return "", fmt.Errorf("found more than one event with type: %s, key: %s", evType, evKey)
	}
	return values[0], nil
}
