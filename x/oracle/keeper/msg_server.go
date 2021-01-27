package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/x/oracle/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) RequestData(goCtx context.Context, msg *types.MsgRequestData) (*types.MsgRequestDataResponse, error) {
	return &types.MsgRequestDataResponse{}, nil
}

func (k msgServer) ReportData(goCtx context.Context, msg *types.MsgReportData) (*types.MsgReportDataResponse, error) {
	return &types.MsgReportDataResponse{}, nil
}

func (k msgServer) CreateDataSource(goCtx context.Context, msg *types.MsgCreateDataSource) (*types.MsgCreateDataSourceResponse, error) {
	return &types.MsgCreateDataSourceResponse{}, nil
}

func (k msgServer) EditDataSource(goCtx context.Context, msg *types.MsgEditDataSource) (*types.MsgEditDataSourceResponse, error) {
	return &types.MsgEditDataSourceResponse{}, nil
}

func (k msgServer) CreateOracleScript(goCtx context.Context, msg *types.MsgCreateOracleScript) (*types.MsgCreateOracleScriptResponse, error) {
	return &types.MsgCreateOracleScriptResponse{}, nil
}

func (k msgServer) EditOracleScript(goCtx context.Context, msg *types.MsgEditOracleScript) (*types.MsgEditOracleScriptResponse, error) {
	return &types.MsgEditOracleScriptResponse{}, nil
}

// Activate changes the given validator's status to active. Returns error if the validator is
// already active or was deactivated recently, as specified by InactivePenaltyDuration parameter.
func (k msgServer) Activate(goCtx context.Context, msg *types.MsgActivate) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	err = k.ActivateValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	return &types.MsgActivateResponse{}, nil
}

// AddReporter adds the reporter address to the list of reporters of the given validator.
func (k msgServer) AddReporter(goCtx context.Context, msg *types.MsgAddReporter) (*types.MsgAddReporterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	repAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	if k.IsReporter(ctx, valAddr, repAddr) {
		return nil, sdkerrors.Wrapf(
			types.ErrReporterAlreadyExists, "val: %s, reporter: %s", msg.Validator, msg.Reporter)
	}
	ctx.KVStore(k.storeKey).Set(types.ReporterStoreKey(valAddr, repAddr), []byte{1})
	return &types.MsgAddReporterResponse{}, nil
}

// RemoveReporter removes the reporter address from the list of reporters of the given validator.
func (k msgServer) RemoveReporter(goCtx context.Context, msg *types.MsgRemoveReporter) (*types.MsgRemoveReporterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	repAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	if !k.IsReporter(ctx, valAddr, repAddr) {
		return nil, sdkerrors.Wrapf(
			types.ErrReporterNotFound, "val: %s, addr: %s", msg.Validator, msg.Reporter)
	}
	ctx.KVStore(k.storeKey).Delete(types.ReporterStoreKey(valAddr, repAddr))
	return &types.MsgRemoveReporterResponse{}, nil
}
