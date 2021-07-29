package keeper

import (
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func ExtendedValidatorsRequestToValidatorsRequest(request *telemetrytypes.QueryExtendedValidatorsRequest) *stakingtypes.QueryValidatorsRequest {
	return &stakingtypes.QueryValidatorsRequest{
		Status:     request.GetStatus(),
		Pagination: request.GetPagination(),
	}
}

func ValidatorsResponseToExtendedValidatorsResponse(request *stakingtypes.QueryValidatorsResponse) *telemetrytypes.QueryExtendedValidatorsResponse {
	return &telemetrytypes.QueryExtendedValidatorsResponse{
		Validators: request.GetValidators(),
		Pagination: request.GetPagination(),
	}
}

func ValidatorsToAccounts(validators []stakingtypes.Validator) ([]sdk.AccAddress, error) {
	accs := make([]sdk.AccAddress, len(validators))
	for i, val := range validators {
		var err error
		accs[i], err = sdk.GetFromBech32(val.OperatorAddress, sdk.GetConfig().GetBech32ValidatorAddrPrefix())
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to get val address from bech32: %s", val)
		}

		if err = sdk.VerifyAddressFormat(accs[i]); err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to get val address from bech32: %s", val)
		}
	}
	return accs, nil
}
