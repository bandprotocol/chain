package event

import (
	"encoding/hex"
	"fmt"
	"strconv"

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

// GetEventValuesUint64 returns the list of all uint64 values in the given log with the given type and key.
func GetEventValuesUint64(log sdk.ABCIMessageLog, evType string, evKey string) ([]uint64, error) {
	strs := GetEventValues(log, evType, evKey)

	var res []uint64
	for _, str := range strs {
		value, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, err
		}

		res = append(res, value)
	}

	return res, nil
}

// GetEventValuesBytes returns the list of all bytes values in the given log with the given type and key.
func GetEventValuesBytes(log sdk.ABCIMessageLog, evType string, evKey string) ([][]byte, error) {
	strs := GetEventValues(log, evType, evKey)

	var res [][]byte
	for _, str := range strs {
		value, err := hex.DecodeString(str)
		if err != nil {
			return nil, err
		}

		res = append(res, value)
	}

	return res, nil
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

// GetEventValueUint64 returns the uin64 value in the given log with the given type and key.
func GetEventValueUint64(log sdk.ABCIMessageLog, evType string, evKey string) (uint64, error) {
	str, err := GetEventValue(log, evType, evKey)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// GetEventValueBytes returns the bytes value in the given log with the given type and key.
func GetEventValueBytes(log sdk.ABCIMessageLog, evType string, evKey string) ([]byte, error) {
	str, err := GetEventValue(log, evType, evKey)
	if err != nil {
		return nil, err
	}

	value, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return value, nil
}
