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
	if len(req.SignalDeviations) > int(params.MaxSignals) {
		return nil, types.ErrMaxSignalsExceeded
	}
	if req.Interval < params.MinInterval {
		return nil, types.ErrMinIntervalExceeded
	}

	creator, err := sdk.AccAddressFromBech32(req.Creator)
	if err != nil {
		return nil, err
	}

	// Add a new tunnel
	tunnel, err := ms.Keeper.AddTunnel(
		ctx,
		req.Route,
		req.Encoder,
		req.SignalDeviations,
		req.Interval,
		creator,
	)
	if err != nil {
		return nil, err
	}

	// Deposit the initial deposit to the tunnel
	if err := ms.Keeper.AddDeposit(ctx, tunnel.ID, creator, req.InitialDeposit); err != nil {
		return nil, err
	}

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
		sdk.NewAttribute(types.AttributeKeyCreator, tunnel.Creator),
	)
	for _, sd := range req.SignalDeviations {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeySignalPriceInfos, sd.String()),
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
	if len(req.SignalDeviations) > int(params.MaxSignals) {
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

	err = ms.Keeper.EditTunnel(ctx, req.TunnelID, req.SignalDeviations, req.Interval)
	if err != nil {
		return nil, err
	}

	return &types.MsgEditTunnelResponse{}, nil
}

// Activate activates a tunnel.
func (ms msgServer) Activate(
	goCtx context.Context,
	req *types.MsgActivate,
) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	// Check if the creator is the same
	if req.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", req.Creator, req.TunnelID)
	}

	// Check if the tunnel is already active
	if tunnel.IsActive {
		return nil, types.ErrAlreadyActive.Wrapf("tunnelID %d", req.TunnelID)
	}

	// verify if the total deposit meets or exceeds the minimum required deposit
	minDeposit := ms.Keeper.GetParams(ctx).MinDeposit
	if tunnel.TotalDeposit.IsAllLT(minDeposit) {
		return nil, types.ErrInsufficientDeposit
	}

	err = ms.Keeper.ActivateTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	return &types.MsgActivateResponse{}, nil
}

// Deactivate deactivates a tunnel.
func (ms msgServer) Deactivate(
	goCtx context.Context,
	req *types.MsgDeactivate,
) (*types.MsgDeactivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	if req.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", req.Creator, req.TunnelID)
	}

	if !tunnel.IsActive {
		return nil, types.ErrAlreadyInactive.Wrapf("tunnelID %d", req.TunnelID)
	}

	err = ms.Keeper.DeactivateTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeactivateResponse{}, nil
}

// TriggerTunnel manually triggers a tunnel.
func (ms msgServer) TriggerTunnel(
	goCtx context.Context,
	req *types.MsgTriggerTunnel,
) (*types.MsgTriggerTunnelResponse, error) {
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

	return &types.MsgTriggerTunnelResponse{}, nil
}

// DepositTunnel adds deposit to the tunnel.
func (ms msgServer) DepositTunnel(
	goCtx context.Context,
	req *types.MsgDepositTunnel,
) (*types.MsgDepositTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, err
	}

	if err := ms.Keeper.AddDeposit(ctx, req.TunnelID, depositor, req.Amount); err != nil {
		return nil, err
	}

	return &types.MsgDepositTunnelResponse{}, nil
}

// WithdrawDepositTunnel withdraws deposit from the tunnel.
func (ms msgServer) WithdrawDepositTunnel(
	goCtx context.Context,
	req *types.MsgWithdrawDepositTunnel,
) (*types.MsgWithdrawDepositTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	withdrawer, err := sdk.AccAddressFromBech32(req.Withdrawer)
	if err != nil {
		return nil, err
	}

	// Withdraw the deposit from the tunnel
	if err := ms.Keeper.WithdrawDeposit(ctx, req.TunnelID, req.Amount, withdrawer); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawDepositTunnelResponse{}, nil
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
