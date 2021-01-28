package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ RequestSpec = &MsgRequestData{}
	_ RequestSpec = &OracleRequestPacketData{}
)

// RequestSpec captures the essence of what it means to be a request-making object.
type RequestSpec interface {
	GetOracleScriptID() OracleScriptID
	GetCalldata() []byte
	GetAskCount() uint64
	GetMinCount() uint64
	GetClientID() string
}

func NewRawRequest(
	ExternalID ExternalID,
	DataSourceID DataSourceID,
	Calldata []byte,
) RawRequest {
	return RawRequest{
		ExternalID:   ExternalID,
		DataSourceID: DataSourceID,
		Calldata:     Calldata,
	}
}

func NewRequest(
	OracleScriptID OracleScriptID,
	Calldata []byte,
	RequestedValidators []sdk.ValAddress,
	MinCount uint64,
	RequestHeight int64,
	RequestTime time.Time,
	ClientID string,
	RawRequests []RawRequest,
) Request {
	requestedVals := make([]string, len(RequestedValidators))
	if RequestedValidators != nil {
		for idx, reqVal := range RequestedValidators {
			requestedVals[idx] = reqVal.String()
		}
	} else {
		requestedVals = nil
	}
	return Request{
		OracleScriptID:      OracleScriptID,
		Calldata:            Calldata,
		RequestedValidators: requestedVals,
		MinCount:            MinCount,
		RequestHeight:       RequestHeight,
		RequestTime:         uint64(RequestTime.UnixNano()),
		ClientID:            ClientID,
		RawRequests:         RawRequests,
	}
}
