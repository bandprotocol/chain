package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the x/tunnel MsgServer interface.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// CreateTunnel creates a new tunnel.
func (ms msgServer) CreateTunnel(
	goCtx context.Context,
	req *types.MsgCreateTunnel,
) (*types.MsgCreateTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check deposit with params, transfer deposit to module account

	var signalPriceInfos []types.SignalPriceInfo
	for _, signalInfo := range req.SignalInfos {
		signalPriceInfos = append(signalPriceInfos, types.SignalPriceInfo{
			SignalID:         signalInfo.SignalID,
			SoftDeviationBPS: signalInfo.SoftDeviationBPS,
			HardDeviationBPS: signalInfo.HardDeviationBPS,
		})
	}

	tunnelID, err := ms.Keeper.AddTunnel(ctx, types.Tunnel{
		Route:            req.Route,
		FeedType:         req.FeedType,
		SignalPriceInfos: signalPriceInfos,
		Interval:         req.Interval,
		IsActive:         false,
		Creator:          req.Creator,
	})
	if err != nil {
		return nil, err
	}

	tunnel, err := ms.Keeper.GetTunnel(ctx, tunnelID)
	if err != nil {
		return nil, err
	}

	event := sdk.NewEvent(
		types.EventTypeCreateTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnel.ID)),
		sdk.NewAttribute(types.AttributeKeyRoute, tunnel.Route.String()),
		sdk.NewAttribute(types.AttributeKeyFeedType, tunnel.FeedType.String()),
		sdk.NewAttribute(types.AttributeKeyFeePayer, tunnel.FeePayer),
		sdk.NewAttribute(types.AttributeKeyIsActive, fmt.Sprintf("%t", tunnel.IsActive)),
		sdk.NewAttribute(types.AttributeKeyCreatedAt, fmt.Sprintf("%d", tunnel.CreatedAt)),
		sdk.NewAttribute(types.AttributeKeyCreator, req.Creator),
	)
	for _, signalInfo := range req.SignalInfos {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeySignalPriceInfos, signalInfo.String()),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateTunnelResponse{
		TunnelID: tunnel.ID,
	}, nil
}

// ActivateTunnel activates a tunnel.
func (ms msgServer) ActivateTunnel(
	goCtx context.Context,
	req *types.MsgActivateTunnel,
) (*types.MsgActivateTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ActivateTunnel(ctx, req.TunnelID, req.Creator)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeActivateTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", req.TunnelID)),
		sdk.NewAttribute(types.AttributeKeyIsActive, fmt.Sprintf("%t", true)),
	))

	return &types.MsgActivateTunnelResponse{}, nil
}

// ManualTriggerTunnel manually triggers a tunnel.
func (ms msgServer) ManualTriggerTunnel(
	goCtx context.Context,
	req *types.MsgManualTriggerTunnel,
) (*types.MsgManualTriggerTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}
	if req.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf(
			"creator %s, tunnelID %d",
			req.Creator,
			req.TunnelID,
		)
	}

	// Add the tunnel to the pending trigger list
	ms.Keeper.AddPendingTriggerTunnel(ctx, req.TunnelID)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeManualTriggerTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", req.TunnelID)),
	))

	return &types.MsgManualTriggerTunnelResponse{}, nil
}

// UpdateParams updates the module params.
func (ms msgServer) UpdateParams(
	goCtx context.Context,
	req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if ms.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			ms.authority,
			req.Authority,
		)
	}

	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, req.Params.String()),
	))

	return &types.MsgUpdateParamsResponse{}, nil
}
