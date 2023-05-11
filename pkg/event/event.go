package event

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GetMessageLogs returns the list of logs from transaction result.
func GetMessageLogs(tx abci.TxResult) (sdk.ABCIMessageLogs, error) {
	if tx.Result.Code != 0 {
		return nil, fmt.Errorf("transaction with non-zero code: %d", tx.Result.Code)
	}

	logs, err := sdk.ParseABCILogs(tx.Result.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction logs with error: %s", err.Error())
	}

	return logs, nil
}

// GetEventValues returns the list of all values in the given log with the given type and key.
func GetEventValues(log sdk.ABCIMessageLog, evType string, evKey string) (res []string) {
	for _, ev := range log.Events {
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
func GetEventValue(log sdk.ABCIMessageLog, evType string, evKey string) (string, error) {
	values := GetEventValues(log, evType, evKey)
	if len(values) == 0 {
		return "", fmt.Errorf("Cannot find event with type: %s, key: %s", evType, evKey)
	}
	if len(values) > 1 {
		return "", fmt.Errorf("Found more than one event with type: %s, key: %s", evType, evKey)
	}
	return values[0], nil
}
