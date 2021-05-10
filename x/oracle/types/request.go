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
	GetPrepareGas() uint64
	GetExecuteGas() uint64
	GetFeeLimit() sdk.Coins
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
	oracleScriptID OracleScriptID,
	calldata []byte,
	requestedValidators []sdk.ValAddress,
	minCount uint64,
	requestHeight int64,
	requestTime time.Time,
	clientID string,
	rawRequests []RawRequest,
	ibcChannel *IBCChannel,
	executeGas uint64,
) Request {
	requestedVals := make([]string, len(requestedValidators))
	if requestedValidators != nil {
		for idx, reqVal := range requestedValidators {
			requestedVals[idx] = reqVal.String()
		}
	} else {
		requestedVals = nil
	}
	return Request{
		OracleScriptID:      oracleScriptID,
		Calldata:            calldata,
		RequestedValidators: requestedVals,
		MinCount:            minCount,
		RequestHeight:       requestHeight,
		RequestTime:         uint64(requestTime.Unix()),
		ClientID:            clientID,
		RawRequests:         rawRequests,
		IBCChannel:          ibcChannel,
		ExecuteGas:          executeGas,
	}
}
