package keeper

import (
	"context"
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
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
	msg *types.MsgCreateTunnel,
) (*types.MsgCreateTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate signal infos and interval
	params := ms.Keeper.GetParams(ctx)
	if len(msg.SignalDeviations) > int(params.MaxSignals) {
		return nil, types.ErrMaxSignalsExceeded
	}

	if msg.Interval < params.MinInterval {
		return nil, types.ErrIntervalTooLow
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	route, ok := msg.Route.GetCachedValue().(types.RouteI)
	if !ok {
		return nil, errors.New("cannot create tunnel, failed to convert proto Any to routeI")
	}

	// add a new tunnel
	tunnel, err := ms.Keeper.AddTunnel(
		ctx,
		route,
		msg.Encoder,
		msg.SignalDeviations,
		msg.Interval,
		creator,
	)
	if err != nil {
		return nil, err
	}

	// Deposit the initial deposit to the tunnel
	if !msg.InitialDeposit.IsZero() {
		if err := ms.Keeper.DepositToTunnel(ctx, tunnel.ID, creator, msg.InitialDeposit); err != nil {
			return nil, err
		}
	}

	return &types.MsgCreateTunnelResponse{
		TunnelID: tunnel.ID,
	}, nil
}

// UpdateAndResetTunnel edits a tunnel and reset latest price interval.
func (ms msgServer) UpdateAndResetTunnel(
	goCtx context.Context,
	msg *types.MsgUpdateAndResetTunnel,
) (*types.MsgUpdateAndResetTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate signal infos and interval
	params := ms.Keeper.GetParams(ctx)
	if len(msg.SignalDeviations) > int(params.MaxSignals) {
		return nil, types.ErrMaxSignalsExceeded
	}
	if msg.Interval < params.MinInterval {
		return nil, types.ErrIntervalTooLow
	}

	tunnel, err := ms.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	err = ms.Keeper.UpdateAndResetTunnel(ctx, msg.TunnelID, msg.SignalDeviations, msg.Interval)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateAndResetTunnelResponse{}, nil
}

// Activate activates a tunnel.
func (ms msgServer) Activate(
	goCtx context.Context,
	msg *types.MsgActivate,
) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	// check if the creator is the same
	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	// check if the tunnel is already active
	if tunnel.IsActive {
		return nil, types.ErrAlreadyActive.Wrapf("tunnelID %d", msg.TunnelID)
	}

	if err := ms.Keeper.ActivateTunnel(ctx, msg.TunnelID); err != nil {
		return nil, err
	}

	return &types.MsgActivateResponse{}, nil
}

// Deactivate deactivates a tunnel.
func (ms msgServer) Deactivate(
	goCtx context.Context,
	msg *types.MsgDeactivate,
) (*types.MsgDeactivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	if !tunnel.IsActive {
		return nil, types.ErrAlreadyInactive.Wrapf("tunnelID %d", msg.TunnelID)
	}

	if err := ms.Keeper.DeactivateTunnel(ctx, msg.TunnelID); err != nil {
		return nil, err
	}

	return &types.MsgDeactivateResponse{}, nil
}

// TriggerTunnel manually triggers a tunnel.
func (ms msgServer) TriggerTunnel(
	goCtx context.Context,
	msg *types.MsgTriggerTunnel,
) (*types.MsgTriggerTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := ms.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf(
			"creator %s, tunnelID %d",
			msg.Creator,
			msg.TunnelID,
		)
	}

	if !tunnel.IsActive {
		return nil, types.ErrInactiveTunnel.Wrapf("tunnelID %d", msg.TunnelID)
	}

	ok, err := ms.Keeper.HasEnoughFundToCreatePacket(ctx, tunnel.ID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, types.ErrInsufficientFund.Wrapf("tunnelID %d", msg.TunnelID)
	}

	signalIDs := tunnel.GetSignalIDs()
	prices := ms.Keeper.feedsKeeper.GetPrices(ctx, signalIDs)

	// create a new packet
	packet, err := ms.CreatePacket(ctx, tunnel.ID, prices)
	if err != nil {
		return nil, err
	}

	// send packet
	if err := ms.SendPacket(ctx, packet); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to create packet for tunnel %d", tunnel.ID)
	}

	latestPrices, err := ms.GetLatestPrices(ctx, tunnel.ID)
	if err != nil {
		return nil, err
	}

	// update latest price info.
	latestPrices.LastInterval = ctx.BlockTime().Unix()
	latestPrices.UpdatePrices(packet.Prices)
	ms.SetLatestPrices(ctx, latestPrices)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTriggerTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", packet.TunnelID)),
		sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", packet.Sequence)),
	))

	return &types.MsgTriggerTunnelResponse{}, nil
}

// DepositToTunnel adds deposit to the tunnel.
func (ms msgServer) DepositToTunnel(
	goCtx context.Context,
	msg *types.MsgDepositToTunnel,
) (*types.MsgDepositToTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}

	if err := ms.Keeper.DepositToTunnel(ctx, msg.TunnelID, depositor, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgDepositToTunnelResponse{}, nil
}

// WithdrawFromTunnel withdraws deposit from the tunnel.
func (ms msgServer) WithdrawFromTunnel(
	goCtx context.Context,
	msg *types.MsgWithdrawFromTunnel,
) (*types.MsgWithdrawFromTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	withdrawer, err := sdk.AccAddressFromBech32(msg.Withdrawer)
	if err != nil {
		return nil, err
	}

	// Withdraw the deposit from the tunnel
	if err := ms.Keeper.WithdrawFromTunnel(ctx, msg.TunnelID, msg.Amount, withdrawer); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawFromTunnelResponse{}, nil
}

// UpdateParams updates the module params.
func (ms msgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if ms.authority != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			ms.authority,
			msg.Authority,
		)
	}

	if err := ms.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, msg.Params.String()),
	))

	return &types.MsgUpdateParamsResponse{}, nil
}
