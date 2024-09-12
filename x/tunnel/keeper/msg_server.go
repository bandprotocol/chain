package keeper

import (
	"context"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
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

	// validate signal infos and interval
	params := ms.Keeper.GetParams(ctx)
	if len(req.SignalInfos) > int(params.MaxSignals) {
		return nil, types.ErrMaxSignalsExceeded
	}
	if req.Interval < params.MinInterval {
		return nil, types.ErrMinIntervalExceeded
	}

	// Get the next tunnel ID
	id := ms.Keeper.GetTunnelCount(ctx)
	newID := id + 1

	// Generate a new fee payer account
	feePayer, err := ms.Keeper.GenerateAccount(ctx, fmt.Sprintf("%d", newID))
	if err != nil {
		return nil, err
	}

	// TODO: check deposit with params, transfer deposit to module account

	// Add a new tunnel
	tunnel := ms.Keeper.AddTunnel(
		ctx,
		newID,
		req.Route,
		req.Encoder,
		feePayer,
		req.SignalInfos,
		req.Interval,
		req.Creator,
	)

	// Increment the tunnel count
	ms.Keeper.SetTunnelCount(ctx, newID)

	// Emit an event
	event := sdk.NewEvent(
		types.EventTypeCreateTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnel.ID)),
		sdk.NewAttribute(types.AttributeKeyInterval, fmt.Sprintf("%d", tunnel.Interval)),
		sdk.NewAttribute(types.AttributeKeyRoute, tunnel.Route.String()),
		sdk.NewAttribute(types.AttributeKeyEncoder, tunnel.Encoder.String()),
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

// EditTunnel edits a tunnel.
func (ms msgServer) EditTunnel(
	goCtx context.Context,
	req *types.MsgEditTunnel,
) (*types.MsgEditTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate signal infos and interval
	params := ms.Keeper.GetParams(ctx)
	if len(req.SignalInfos) > int(params.MaxSignals) {
		return nil, types.ErrMaxSignalsExceeded
	}
	if req.Interval < params.MinInterval {
		return nil, types.ErrMinIntervalExceeded
	}

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	if req.Creator != tunnel.Creator {
		return nil, fmt.Errorf("creator %s is not the creator of tunnel %d", req.Creator, req.TunnelID)
	}

	err = ms.Keeper.EditTunnel(ctx, req.TunnelID, req.SignalInfos, req.Interval)
	if err != nil {
		return nil, err
	}

	// Emit an event
	event := sdk.NewEvent(
		types.EventTypeEditTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnel.ID)),
		sdk.NewAttribute(types.AttributeKeyInterval, fmt.Sprintf("%d", tunnel.Interval)),
		sdk.NewAttribute(types.AttributeKeyCreator, req.Creator),
	)
	for _, signalInfo := range req.SignalInfos {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeySignalPriceInfos, signalInfo.String()),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgEditTunnelResponse{}, nil
}

// ActivateTunnel activates a tunnel.
func (ms msgServer) ActivateTunnel(
	goCtx context.Context,
	req *types.MsgActivateTunnel,
) (*types.MsgActivateTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	if req.Creator != tunnel.Creator {
		return nil, fmt.Errorf("creator %s is not the creator of tunnel %d", req.Creator, req.TunnelID)
	}

	err = ms.Keeper.ActivateTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	return &types.MsgActivateTunnelResponse{}, nil
}

// DeactivateTunnel deactivates a tunnel.
func (ms msgServer) DeactivateTunnel(
	goCtx context.Context,
	req *types.MsgDeactivateTunnel,
) (*types.MsgDeactivateTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	if req.Creator != tunnel.Creator {
		return nil, fmt.Errorf("creator %s is not the creator of tunnel %d", req.Creator, req.TunnelID)
	}

	err = ms.Keeper.DeactivateTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeactivateTunnelResponse{}, nil
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
	if !tunnel.IsActive {
		return nil, types.ErrInactiveTunnel.Wrapf("tunnelID %d", req.TunnelID)
	}

	currentPrices := ms.Keeper.feedsKeeper.GetCurrentPrices(ctx)
	currentPricesMap := createCurrentPricesMap(currentPrices)

	// Produce packet with trigger all signals
	isCreated, err := ms.Keeper.ProducePacket(ctx, tunnel.ID, currentPricesMap, true)
	if err != nil {
		return nil, err
	}

	// if new packet is created, deduct base packet fee from the fee payer,
	if isCreated {
		feePayer, err := sdk.AccAddressFromBech32(tunnel.FeePayer)
		if err != nil {
			return nil, err
		}

		if err := ms.Keeper.DeductBasePacketFee(ctx, feePayer); err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to deduct base packet fee for tunnel %d", req.TunnelID)
		}
	}

	// Emit an event
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

	// Emit an event
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, req.Params.String()),
	))

	return &types.MsgUpdateParamsResponse{}, nil
}
