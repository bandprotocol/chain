package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewRequestVerification(
	chainID string,
	validator sdk.ValAddress,
	requestID RequestID,
	externalID ExternalID,
	dataSourceID DataSourceID,
) RequestVerification {
	return RequestVerification{
		ChainID:      chainID,
		Validator:    validator.String(),
		RequestID:    requestID,
		ExternalID:   externalID,
		DataSourceID: dataSourceID,
	}
}

func (msg RequestVerification) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}
