package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/tunnel MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
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
		return nil, types.ErrIntervalTooLow
	}

	creator, err := sdk.AccAddressFromBech32(req.Creator)
	if err != nil {
		return nil, err
	}

	route, ok := req.Route.GetCachedValue().(types.RouteI)
	if !ok {
		return nil, errors.New("cannot create tunnel, failed to convert proto Any to routeI")
	}

	// add a new tunnel
	tunnel, err := ms.Keeper.AddTunnel(
		ctx,
		route,
		req.Encoder,
		req.SignalDeviations,
		req.Interval,
		creator,
	)
	if err != nil {
		return nil, err
	}

	// Deposit the initial deposit to the tunnel
	if !req.InitialDeposit.IsZero() {
		if err := ms.Keeper.AddDeposit(ctx, tunnel.ID, creator, req.InitialDeposit); err != nil {
			return nil, err
		}
	}

	// Emit an event
	event := sdk.NewEvent(
		types.EventTypeCreateTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnel.ID)),
		sdk.NewAttribute(types.AttributeKeyInterval, fmt.Sprintf("%d", tunnel.Interval)),
		sdk.NewAttribute(types.AttributeKeyRoute, tunnel.Route.String()),
		sdk.NewAttribute(types.AttributeKeyEncoder, tunnel.Encoder.String()),
		sdk.NewAttribute(types.AttributeKeyInitialDeposit, req.InitialDeposit.String()),
		sdk.NewAttribute(types.AttributeKeyFeePayer, tunnel.FeePayer),
		sdk.NewAttribute(types.AttributeKeyIsActive, fmt.Sprintf("%t", tunnel.IsActive)),
		sdk.NewAttribute(types.AttributeKeyCreatedAt, fmt.Sprintf("%d", tunnel.CreatedAt)),
		sdk.NewAttribute(types.AttributeKeyCreator, tunnel.Creator),
	)
	for _, sd := range req.SignalDeviations {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeySignalDeviation, sd.String()),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateTunnelResponse{
		TunnelID: tunnel.ID,
	}, nil
}

// UpdateAndResetTunnel edits a tunnel and reset latest signal price interval.
func (ms msgServer) UpdateAndResetTunnel(
	goCtx context.Context,
	req *types.MsgUpdateAndResetTunnel,
) (*types.MsgUpdateAndResetTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate signal infos and interval
	params := ms.Keeper.GetParams(ctx)
	if len(req.SignalDeviations) > int(params.MaxSignals) {
		return nil, types.ErrMaxSignalsExceeded
	}
	if req.Interval < params.MinInterval {
		return nil, types.ErrIntervalTooLow
	}

	tunnel, err := ms.Keeper.GetTunnel(ctx, req.TunnelID)
	if err != nil {
		return nil, err
	}

	if req.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", req.Creator, req.TunnelID)
	}

	err = ms.Keeper.UpdateAndResetTunnel(ctx, req.TunnelID, req.SignalDeviations, req.Interval)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateAndResetTunnelResponse{}, nil
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

	// check if the creator is the same
	if req.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", req.Creator, req.TunnelID)
	}

	// check if the tunnel is already active
	if tunnel.IsActive {
		return nil, types.ErrAlreadyActive.Wrapf("tunnelID %d", req.TunnelID)
	}

	if err := ms.Keeper.ActivateTunnel(ctx, req.TunnelID); err != nil {
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

	if err := ms.Keeper.DeactivateTunnel(ctx, req.TunnelID); err != nil {
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

	ok, err := ms.Keeper.HasEnoughFundToCreatePacket(ctx, tunnel.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, types.ErrInsufficientFund.Wrapf("tunnelID %d", req.TunnelID)
	}

	signalIDs := tunnel.GetSignalIDs()
	currentPrices := ms.Keeper.feedsKeeper.GetCurrentPrices(ctx, signalIDs)
	currentPricesMap := createPricesMap(currentPrices)

	if err := ms.Keeper.ProducePacket(ctx, req.TunnelID, currentPricesMap, true); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTriggerTunnel,
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

// WithdrawTunnel withdraws deposit from the tunnel.
func (ms msgServer) WithdrawTunnel(
	goCtx context.Context,
	req *types.MsgWithdrawTunnel,
) (*types.MsgWithdrawTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	withdrawer, err := sdk.AccAddressFromBech32(req.Withdrawer)
	if err != nil {
		return nil, err
	}

	// Withdraw the deposit from the tunnel
	if err := ms.Keeper.WithdrawDeposit(ctx, req.TunnelID, req.Amount, withdrawer); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawTunnelResponse{}, nil
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
