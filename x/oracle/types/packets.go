package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewOracleRequestPacketData contructs a new OracleRequestPacketData instance
func NewOracleRequestPacketData(
	clientID string, oracleScriptID OracleScriptID, calldata []byte, askCount uint64, minCount uint64, feeLimit sdk.Coins, requestKey string,
) OracleRequestPacketData {
	return OracleRequestPacketData{
		ClientID:       clientID,
		OracleScriptID: oracleScriptID,
		Calldata:       calldata,
		AskCount:       askCount,
		MinCount:       minCount,
		FeeLimit:       feeLimit,
		RequestKey:     requestKey,
	}
}

// ValidateBasic is used for validating the request.
func (p OracleRequestPacketData) ValidateBasic() error {
	// not checking for max data here, will check it in handler
	if p.MinCount <= 0 {
		return sdkerrors.Wrapf(ErrInvalidMinCount, "got: %d", p.MinCount)
	}
	if p.AskCount < p.MinCount {
		return sdkerrors.Wrapf(ErrInvalidAskCount, "got: %d, min count: %d", p.AskCount, p.MinCount)
	}
	if len(p.ClientID) > MaxClientIDLength {
		return WrapMaxError(ErrTooLongClientID, len(p.ClientID), MaxClientIDLength)
	}
	return nil
}

// GetBytes is a helper for serialising
func (p OracleRequestPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}

func NewOracleRequestPacketAcknowledgement(requestID RequestID) *OracleRequestPacketAcknowledgement {
	return &OracleRequestPacketAcknowledgement{
		RequestID: requestID,
	}
}

// NewOracleResponsePacketData contructs a new OracleResponsePacketData instance
func NewOracleResponsePacketData(
	clientID string, requestID RequestID, ansCount uint64, requestTime int64,
	resolveTime int64, resolveStatus ResolveStatus, result []byte,
) OracleResponsePacketData {
	return OracleResponsePacketData{
		ClientID:      clientID,
		RequestID:     requestID,
		AnsCount:      ansCount,
		RequestTime:   requestTime,
		ResolveTime:   resolveTime,
		ResolveStatus: resolveStatus,
		Result:        result,
	}
}

// GetBytes returns the bytes representation of this oracle response packet data.
func (p OracleResponsePacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
