package oracle

import (
	"github.com/bandprotocol/chain/x/oracle/keeper"
	"github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates the msg handler of this module, as required by Cosmos-SDK standard.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		msgServer := keeper.NewMsgServerImpl(k)
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgRequestData:
			res, err := msgServer.RequestData(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgReportData:
			res, err := msgServer.ReportData(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCreateDataSource:
			res, err := msgServer.CreateDataSource(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgEditDataSource:
			res, err := msgServer.EditDataSource(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCreateOracleScript:
			res, err := msgServer.CreateOracleScript(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgEditOracleScript:
			res, err := msgServer.EditOracleScript(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgActivate:
			res, err := msgServer.Activate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddReporter:
			res, err := msgServer.AddReporter(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRemoveReporter:
			res, err := msgServer.RemoveReporter(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

// func handleMsgCreateDataSource(ctx sdk.Context, k Keeper, m MsgCreateDataSource) (*sdk.Result, error) {
// 	if gzip.IsGzipped(m.Executable) {
// 		var err error
// 		m.Executable, err = gzip.Uncompress(m.Executable, types.MaxExecutableSize)
// 		if err != nil {
// 			return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 		}
// 	}
// 	owner, err := sdk.AccAddressFromBech32(m.Owner)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	id := k.AddDataSource(ctx, types.NewDataSource(
// 		owner, m.Name, m.Description, k.AddExecutableFile(m.Executable),
// 	))
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeCreateDataSource,
// 		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgEditDataSource(ctx sdk.Context, k Keeper, m MsgEditDataSource) (*sdk.Result, error) {
// 	dataSource, err := k.GetDataSource(ctx, m.DataSourceID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	owner, err := sdk.AccAddressFromBech32(dataSource.Owner)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	sender, err := sdk.AccAddressFromBech32(m.Sender)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	if !owner.Equals(sender) {
// 		return nil, types.ErrEditorNotAuthorized
// 	}
// 	if gzip.IsGzipped(m.Executable) {
// 		m.Executable, err = gzip.Uncompress(m.Executable, types.MaxExecutableSize)
// 		if err != nil {
// 			return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 		}
// 	}
// 	// Can safely use MustEdit here, as we already checked that the data source exists above.
// 	k.MustEditDataSource(ctx, m.DataSourceID, types.NewDataSource(
// 		owner, m.Name, m.Description, k.AddExecutableFile(m.Executable),
// 	))
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeEditDataSource,
// 		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", m.DataSourceID)),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgCreateOracleScript(ctx sdk.Context, k Keeper, m MsgCreateOracleScript) (*sdk.Result, error) {
// 	if gzip.IsGzipped(m.Code) {
// 		var err error
// 		m.Code, err = gzip.Uncompress(m.Code, types.MaxWasmCodeSize)
// 		if err != nil {
// 			return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 		}
// 	}
// 	owner, err := sdk.AccAddressFromBech32(m.Owner)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	filename, err := k.AddOracleScriptFile(m.Code)
// 	if err != nil {
// 		return nil, err
// 	}
// 	id := k.AddOracleScript(ctx, types.NewOracleScript(
// 		owner, m.Name, m.Description, filename, m.Schema, m.SourceCodeURL,
// 	))
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeCreateOracleScript,
// 		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgEditOracleScript(ctx sdk.Context, k Keeper, m MsgEditOracleScript) (*sdk.Result, error) {
// 	oracleScript, err := k.GetOracleScript(ctx, m.OracleScriptID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	owner, err := sdk.AccAddressFromBech32(oracleScript.Owner)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	sender, err := sdk.AccAddressFromBech32(oracleScript.Owner)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	if !owner.Equals(sender) {
// 		return nil, types.ErrEditorNotAuthorized
// 	}
// 	if gzip.IsGzipped(m.Code) {
// 		m.Code, err = gzip.Uncompress(m.Code, types.MaxWasmCodeSize)
// 		if err != nil {
// 			return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 		}
// 	}
// 	filename, err := k.AddOracleScriptFile(m.Code)
// 	if err != nil {
// 		return nil, err
// 	}
// 	k.MustEditOracleScript(ctx, m.OracleScriptID, types.NewOracleScript(
// 		owner, m.Name, m.Description, filename, m.Schema, m.SourceCodeURL,
// 	))
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeEditOracleScript,
// 		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", m.OracleScriptID)),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgRequestData(ctx sdk.Context, k Keeper, m MsgRequestData) (*sdk.Result, error) {
// 	err := k.PrepareRequest(ctx, &m)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgReportData(ctx sdk.Context, k Keeper, m MsgReportData) (*sdk.Result, error) {
// 	validator, err := sdk.ValAddressFromBech32(m.Validator)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	reporter, err := sdk.AccAddressFromBech32(m.Reporter)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	if !k.IsReporter(ctx, validator, reporter) {
// 		return nil, types.ErrReporterNotAuthorized
// 	}
// 	if m.RequestID <= k.GetRequestLastExpired(ctx) {
// 		return nil, types.ErrRequestAlreadyExpired
// 	}
// 	err = k.AddReport(ctx, m.RequestID, types.NewReport(validator, !k.HasResult(ctx, m.RequestID), m.RawReports))
// 	if err != nil {
// 		return nil, err
// 	}
// 	req := k.MustGetRequest(ctx, m.RequestID)
// 	if k.GetReportCount(ctx, m.RequestID) == req.MinCount {
// 		// At the exact moment when the number of reports is sufficient, we add the request to
// 		// the pending resolve list. This can happen at most one time for any request.
// 		k.AddPendingRequest(ctx, m.RequestID)
// 	}
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeReport,
// 		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", m.RequestID)),
// 		sdk.NewAttribute(types.AttributeKeyValidator, m.Validator),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgActivate(ctx sdk.Context, k Keeper, m MsgActivate) (*sdk.Result, error) {
// 	validator, err := sdk.ValAddressFromBech32(m.Validator)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	err = k.Activate(ctx, validator)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeActivate,
// 		sdk.NewAttribute(types.AttributeKeyValidator, m.Validator),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgAddReporter(ctx sdk.Context, k Keeper, m MsgAddReporter) (*sdk.Result, error) {
// 	validator, err := sdk.ValAddressFromBech32(m.Validator)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	reporter, err := sdk.AccAddressFromBech32(m.Reporter)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	err = k.AddReporter(ctx, validator, reporter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeAddReporter,
// 		sdk.NewAttribute(types.AttributeKeyValidator, m.Validator),
// 		sdk.NewAttribute(types.AttributeKeyReporter, m.Reporter),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

// func handleMsgRemoveReporter(ctx sdk.Context, k Keeper, m MsgRemoveReporter) (*sdk.Result, error) {
// 	validator, err := sdk.ValAddressFromBech32(m.Validator)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	reporter, err := sdk.AccAddressFromBech32(m.Reporter)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(types.ErrUncompressionFailed, err.Error())
// 	}
// 	err = k.RemoveReporter(ctx, validator, reporter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ctx.EventManager().EmitEvent(sdk.NewEvent(
// 		types.EventTypeRemoveReporter,
// 		sdk.NewAttribute(types.AttributeKeyValidator, m.Validator),
// 		sdk.NewAttribute(types.AttributeKeyReporter, m.Reporter),
// 	))
// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }
