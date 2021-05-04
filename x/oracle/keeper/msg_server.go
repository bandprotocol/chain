package oraclekeeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/GeoDB-Limited/odin-core/pkg/gzip"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the oracle MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) oracletypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ oracletypes.MsgServer = msgServer{}

func (k msgServer) RequestData(goCtx context.Context, msg *oracletypes.MsgRequestData) (*oracletypes.MsgRequestDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	maxCalldataSize := k.GetParamUint64(ctx, oracletypes.KeyMaxCalldataSize)
	if len(msg.Calldata) > int(maxCalldataSize) {
		return nil, oracletypes.WrapMaxError(oracletypes.ErrTooLargeCalldata, len(msg.Calldata), int(maxCalldataSize))
	}

	payer, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = k.PrepareRequest(ctx, msg, payer, nil)
	if err != nil {
		return nil, err
	}
	return &oracletypes.MsgRequestDataResponse{}, nil
}

func (k msgServer) ReportData(goCtx context.Context, msg *oracletypes.MsgReportData) (*oracletypes.MsgReportDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	reporter, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, err
	}

	// check this address is a reporter of the validator
	if !k.IsReporter(ctx, validator, reporter) {
		return nil, oracletypes.ErrReporterNotAuthorized
	}

	// check request must not expire.
	if msg.RequestID <= k.GetRequestLastExpired(ctx) {
		return nil, oracletypes.ErrRequestAlreadyExpired
	}

	maxDataSize := k.GetParamUint64(ctx, oracletypes.KeyMaxDataSize)
	for _, r := range msg.RawReports {
		if len(r.Data) > int(maxDataSize) {
			return nil, oracletypes.WrapMaxError(oracletypes.ErrTooLargeRawReportData, len(r.Data), int(maxDataSize))
		}
	}

	reportInTime := !k.HasResult(ctx, msg.RequestID)
	err = k.AddReport(ctx, msg.RequestID, oracletypes.NewReport(validator, reportInTime, msg.RawReports))
	if err != nil {
		return nil, err
	}

	// if request has not been resolved, check if it need to resolve at the endblock
	if reportInTime {
		req := k.MustGetRequest(ctx, msg.RequestID)
		if k.GetReportCount(ctx, msg.RequestID) == req.MinCount {
			// at this moment we are sure, that all the raw reports here are validated
			// so we can distribute the reward for them in end-block
			if _, err := k.CollectReward(ctx, msg.GetRawReports(), req.RawRequests); err != nil {
				return nil, err
			}
			// At the exact moment when the number of reports is sufficient, we add the request to
			// the pending resolve list. This can happen at most one time for any request.
			k.AddPendingRequest(ctx, msg.RequestID)
		}
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeReport,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", msg.RequestID)),
		sdk.NewAttribute(oracletypes.AttributeKeyValidator, msg.Validator),
	))
	return &oracletypes.MsgReportDataResponse{}, nil
}

func (k msgServer) CreateDataSource(goCtx context.Context, msg *oracletypes.MsgCreateDataSource) (*oracletypes.MsgCreateDataSourceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// unzip if it's a zip file
	if gzip.IsGzipped(msg.Executable) {
		var err error
		msg.Executable, err = gzip.Uncompress(msg.Executable, oracletypes.MaxExecutableSize)
		if err != nil {
			return nil, sdkerrors.Wrapf(oracletypes.ErrUncompressionFailed, err.Error())
		}
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	id := k.AddDataSource(ctx, oracletypes.NewDataSource(
		owner, msg.Name, msg.Description, k.AddExecutableFile(msg.Executable), msg.Fee,
	))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeCreateDataSource,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", id)),
	))

	return &oracletypes.MsgCreateDataSourceResponse{}, nil
}

func (k msgServer) EditDataSource(goCtx context.Context, msg *oracletypes.MsgEditDataSource) (*oracletypes.MsgEditDataSourceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dataSource, err := k.GetDataSource(ctx, msg.DataSourceID)
	if err != nil {
		return nil, err
	}

	owner, err := sdk.AccAddressFromBech32(dataSource.Owner)
	if err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// sender must be the owner of data source
	if !owner.Equals(sender) {
		return nil, oracletypes.ErrEditorNotAuthorized
	}

	// unzip if it's a zip file
	if gzip.IsGzipped(msg.Executable) {
		msg.Executable, err = gzip.Uncompress(msg.Executable, oracletypes.MaxExecutableSize)
		if err != nil {
			return nil, sdkerrors.Wrapf(oracletypes.ErrUncompressionFailed, err.Error())
		}
	}

	// Can safely use MustEdit here, as we already checked that the data source exists above.
	k.MustEditDataSource(ctx, msg.DataSourceID, oracletypes.NewDataSource(
		owner, msg.Name, msg.Description, k.AddExecutableFile(msg.Executable), msg.Fee,
	))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeEditDataSource,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", msg.DataSourceID)),
	))

	return &oracletypes.MsgEditDataSourceResponse{}, nil
}

func (k msgServer) CreateOracleScript(goCtx context.Context, msg *oracletypes.MsgCreateOracleScript) (*oracletypes.MsgCreateOracleScriptResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// unzip if it's a zip file
	if gzip.IsGzipped(msg.Code) {
		var err error
		msg.Code, err = gzip.Uncompress(msg.Code, oracletypes.MaxWasmCodeSize)
		if err != nil {
			return nil, sdkerrors.Wrapf(oracletypes.ErrUncompressionFailed, err.Error())
		}
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	filename, err := k.AddOracleScriptFile(msg.Code)
	if err != nil {
		return nil, err
	}

	id := k.AddOracleScript(ctx, oracletypes.NewOracleScript(
		owner, msg.Name, msg.Description, filename, msg.Schema, msg.SourceCodeURL,
	))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeCreateOracleScript,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", id)),
	))

	return &oracletypes.MsgCreateOracleScriptResponse{}, nil
}

func (k msgServer) EditOracleScript(goCtx context.Context, msg *oracletypes.MsgEditOracleScript) (*oracletypes.MsgEditOracleScriptResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	oracleScript, err := k.GetOracleScript(ctx, msg.OracleScriptID)
	if err != nil {
		return nil, err
	}

	owner, err := sdk.AccAddressFromBech32(oracleScript.Owner)
	if err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// sender must be the owner of oracle script
	if !owner.Equals(sender) {
		return nil, oracletypes.ErrEditorNotAuthorized
	}

	// unzip if it's a zip file
	if gzip.IsGzipped(msg.Code) {
		msg.Code, err = gzip.Uncompress(msg.Code, oracletypes.MaxWasmCodeSize)
		if err != nil {
			return nil, sdkerrors.Wrapf(oracletypes.ErrUncompressionFailed, err.Error())
		}
	}

	filename, err := k.AddOracleScriptFile(msg.Code)
	if err != nil {
		return nil, err
	}

	k.MustEditOracleScript(ctx, msg.OracleScriptID, oracletypes.NewOracleScript(
		owner, msg.Name, msg.Description, filename, msg.Schema, msg.SourceCodeURL,
	))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeEditOracleScript,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", msg.OracleScriptID)),
	))

	return &oracletypes.MsgEditOracleScriptResponse{}, nil
}

func (k msgServer) Activate(goCtx context.Context, msg *oracletypes.MsgActivate) (*oracletypes.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	err = k.Keeper.Activate(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeActivate,
		sdk.NewAttribute(oracletypes.AttributeKeyValidator, msg.Validator),
	))
	return &oracletypes.MsgActivateResponse{}, nil
}

func (k msgServer) AddReporter(goCtx context.Context, msg *oracletypes.MsgAddReporter) (*oracletypes.MsgAddReporterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	repAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	err = k.Keeper.AddReporter(ctx, valAddr, repAddr)
	if err != nil {
		return nil, err
	}
	ctx.KVStore(k.storeKey).Set(oracletypes.ReporterStoreKey(valAddr, repAddr), []byte{1})
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeAddReporter,
		sdk.NewAttribute(oracletypes.AttributeKeyValidator, msg.Validator),
		sdk.NewAttribute(oracletypes.AttributeKeyReporter, msg.Reporter),
	))
	return &oracletypes.MsgAddReporterResponse{}, nil
}

func (k msgServer) RemoveReporter(goCtx context.Context, msg *oracletypes.MsgRemoveReporter) (*oracletypes.MsgRemoveReporterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	repAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	err = k.Keeper.RemoveReporter(ctx, valAddr, repAddr)
	if err != nil {
		return nil, err
	}
	ctx.KVStore(k.storeKey).Delete(oracletypes.ReporterStoreKey(valAddr, repAddr))
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeRemoveReporter,
		sdk.NewAttribute(oracletypes.AttributeKeyValidator, msg.Validator),
		sdk.NewAttribute(oracletypes.AttributeKeyReporter, msg.Reporter),
	))
	return &oracletypes.MsgRemoveReporterResponse{}, nil
}
