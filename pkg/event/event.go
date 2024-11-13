package event

import (
	"encoding/hex"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEventValues returns the list of all values in the given events with the given type and key.
func GetEventValues(events sdk.StringEvents, evType string, evKey string) (res []string) {
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

// GetEventValuesUint64 returns the list of all uint64 values in the given events with the given type and key.
func GetEventValuesUint64(events sdk.StringEvents, evType string, evKey string) ([]uint64, error) {
	strs := GetEventValues(events, evType, evKey)

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

// GetEventValuesBytes returns the list of all bytes values in the given events with the given type and key.
func GetEventValuesBytes(events sdk.StringEvents, evType string, evKey string) ([][]byte, error) {
	strs := GetEventValues(events, evType, evKey)

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

// GetEventValue checks and returns the exact value in the given events with the given type and key.
func GetEventValue(events sdk.StringEvents, evType string, evKey string) (string, error) {
	values := GetEventValues(events, evType, evKey)
	if len(values) == 0 {
		return "", fmt.Errorf("cannot find event with type: %s, key: %s", evType, evKey)
	}
	if len(values) > 1 {
		return "", fmt.Errorf("found more than one event with type: %s, key: %s", evType, evKey)
	}
	return values[0], nil
}

// GetEventValueUint64 returns the uin64 value in the given events with the given type and key.
func GetEventValueUint64(events sdk.StringEvents, evType string, evKey string) (uint64, error) {
	str, err := GetEventValue(events, evType, evKey)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// GetEventValueBytes returns the bytes value in the given events with the given type and key.
func GetEventValueBytes(events sdk.StringEvents, evType string, evKey string) ([]byte, error) {
	str, err := GetEventValue(events, evType, evKey)
	if err != nil {
		return nil, err
	}

	value, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return value, nil
}
