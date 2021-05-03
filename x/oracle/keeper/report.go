package oraclekeeper

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"
)

// HasReport checks if the report of this ID triple exists in the storage.
func (k Keeper) HasReport(ctx sdk.Context, rid oracletypes.RequestID, val sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has(oracletypes.ReportsOfValidatorPrefixKey(rid, val))
}

// SetReport saves the report to the storage without performing validation.
func (k Keeper) SetReport(ctx sdk.Context, rid oracletypes.RequestID, rep oracletypes.Report) {
	val, _ := sdk.ValAddressFromBech32(rep.Validator)
	key := oracletypes.ReportsOfValidatorPrefixKey(rid, val)
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshalBinaryBare(&rep))
}

// AddReport performs sanity checks and adds a new batch from one validator to one request
// to the store. Note that we expect each validator to report to all raw data requests at once.
func (k Keeper) AddReport(ctx sdk.Context, rid oracletypes.RequestID, rep oracletypes.Report) error {
	req, err := k.GetRequest(ctx, rid)
	if err != nil {
		return err
	}
	val, err := sdk.ValAddressFromBech32(rep.Validator)
	if err != nil {
		return err
	}
	reqVals := make([]sdk.ValAddress, len(req.RequestedValidators))
	for idx, reqVal := range req.RequestedValidators {
		v, err := sdk.ValAddressFromBech32(reqVal)
		if err != nil {
			return err
		}
		reqVals[idx] = v
	}
	if !ContainsVal(reqVals, val) {
		return sdkerrors.Wrapf(
			oracletypes.ErrValidatorNotRequested, "reqID: %d, val: %s", rid, rep.Validator)
	}
	if k.HasReport(ctx, rid, val) {
		return sdkerrors.Wrapf(
			oracletypes.ErrValidatorAlreadyReported, "reqID: %d, val: %s", rid, rep.Validator)
	}
	if len(rep.RawReports) != len(req.RawRequests) {
		return oracletypes.ErrInvalidReportSize
	}
	for _, rep := range rep.RawReports {
		// Here we can safely assume that external IDs are unique, as this has already been
		// checked by ValidateBasic performed in baseapp's runTx function.
		if !ContainsEID(req.RawRequests, rep.ExternalID) {
			return sdkerrors.Wrapf(
				oracletypes.ErrRawRequestNotFound, "reqID: %d, extID: %d", rid, rep.ExternalID)
		}
	}
	k.SetReport(ctx, rid, rep)
	return nil
}

// GetReportIterator returns the iterator for all reports of the given request ID.
func (k Keeper) GetReportIterator(ctx sdk.Context, rid oracletypes.RequestID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), oracletypes.ReportStoreKey(rid))
}

// GetReportCount returns the number of reports for the given request ID.
func (k Keeper) GetReportCount(ctx sdk.Context, rid oracletypes.RequestID) (count uint64) {
	iterator := k.GetReportIterator(ctx, rid)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// GetRequestReports returns all reports for the given request ID, or nil if there is none.
func (k Keeper) GetRequestReports(ctx sdk.Context, rid oracletypes.RequestID) (reports []oracletypes.Report) {
	iterator := k.GetReportIterator(ctx, rid)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var rep oracletypes.Report
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &rep)
		reports = append(reports, rep)
	}
	return reports
}

// GetPaginatedRequestReports returns all reports for the given request ID with pagination.
func (k Keeper) GetPaginatedRequestReports(
	ctx sdk.Context,
	rid oracletypes.RequestID,
	pagination *query.PageRequest,
) ([]oracletypes.Report, *query.PageResponse, error) {
	reports := make([]oracletypes.Report, 0)
	reportsStore := prefix.NewStore(ctx.KVStore(k.storeKey), oracletypes.ReportStoreKey(rid))

	pageRes, err := query.FilteredPaginate(
		reportsStore,
		pagination,
		func(key []byte, value []byte, accumulate bool) (bool, error) {
			var report oracletypes.Report
			if err := k.cdc.UnmarshalBinaryBare(value, &report); err != nil {
				return false, err
			}
			if accumulate {
				reports = append(reports, report)
			}
			return true, nil
		},
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to paginate request reports")
	}

	return reports, pageRes, nil
}

// DeleteReports removes all reports for the given request ID.
func (k Keeper) DeleteReports(ctx sdk.Context, rid oracletypes.RequestID) {
	var keys [][]byte
	iterator := k.GetReportIterator(ctx, rid)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
	}
	for _, key := range keys {
		ctx.KVStore(k.storeKey).Delete(key)
	}
}
