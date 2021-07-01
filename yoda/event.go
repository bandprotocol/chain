package yoda

import (
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

type rawRequest struct {
	dataSource     types.DataSource
	dataSourceID   types.DataSourceID
	externalID     types.ExternalID
	dataSourceHash string
	calldata       string
}

// GetRawRequests returns the list of all raw data requests in the given log.
func GetRawRequests(c *Context, l *Logger, log sdk.ABCIMessageLog) ([]rawRequest, error) {
	dataSourceIDs := GetEventValues(log, types.EventTypeRawRequest, types.AttributeKeyDataSourceID)
	dataSourceHashList := GetEventValues(log, types.EventTypeRawRequest, types.AttributeKeyDataSourceHash)
	externalIDs := GetEventValues(log, types.EventTypeRawRequest, types.AttributeKeyExternalID)
	calldataList := GetEventValues(log, types.EventTypeRawRequest, types.AttributeKeyCalldata)

	if len(dataSourceIDs) != len(externalIDs) {
		return nil, sdkerrors.Wrap(errors.ErrInconsistentCount, "inconsistent data source count and external ID count")
	}
	if len(dataSourceIDs) != len(calldataList) {
		return nil, sdkerrors.Wrap(errors.ErrInconsistentCount, "inconsistent data source count and calldata count")
	}

	var reqs []rawRequest
	for idx := range dataSourceIDs {
		dataSourceID, err := strconv.Atoi(dataSourceIDs[idx])
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to parse data source id")
		}

		externalID, err := strconv.Atoi(externalIDs[idx])
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to parse external id")
		}

		ds, err := GetDataSource(c, l, types.DataSourceID(dataSourceID))
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to get data source by id")
		}

		reqs = append(reqs, rawRequest{
			dataSourceID:   types.DataSourceID(dataSourceID),
			dataSourceHash: dataSourceHashList[idx],
			externalID:     types.ExternalID(externalID),
			calldata:       calldataList[idx],
			dataSource:     ds,
		})
	}

	return reqs, nil
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
		return "", sdkerrors.Wrapf(errors.ErrUnknownEventType, "cannot find event with type: %s, key: %s", evType, evKey)
	}
	if len(values) > 1 {
		return "", sdkerrors.Wrapf(errors.ErrInvalidEventsCount, "found more than one event with type: %s, key: %s", evType, evKey)
	}
	return values[0], nil
}
