package keeper

import (
	"context"
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
func (k msgServer) CreateTunnel(
	goCtx context.Context,
	msg *types.MsgCreateTunnel,
) (*types.MsgCreateTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.Keeper.GetParams(ctx)

	// validate signal infos and interval
	if err := types.ValidateSignalDeviations(msg.SignalDeviations, params.MaxSignals, params.MaxDeviationBPS, params.MinDeviationBPS); err != nil {
		return nil, err
	}

	// validate interval
	if err := types.ValidateInterval(msg.Interval, params.MaxInterval, params.MinInterval); err != nil {
		return nil, err
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	route, err := msg.GetRouteValue()
	if err != nil {
		return nil, err
	}

	// Check channel id in ibc route should be empty
	ibcRoute, isIBCRoute := route.(*types.IBCRoute)
	if isIBCRoute {
		if ibcRoute.ChannelID != "" {
			return nil, types.ErrInvalidRoute.Wrap("channel id should be set after create tunnel")
		}
	}

	// add a new tunnel
	tunnel, err := k.Keeper.AddTunnel(
		ctx,
		route,
		msg.SignalDeviations,
		msg.Interval,
		creator,
	)
	if err != nil {
		return nil, err
	}

	// Bind ibc port for the new tunnel
	if isIBCRoute {
		_, err = k.Keeper.ensureIBCPort(ctx, tunnel.ID)
		if err != nil {
			return nil, err
		}
	}

	// Deposit the initial deposit to the tunnel
	if !msg.InitialDeposit.IsZero() {
		if err := k.Keeper.DepositToTunnel(ctx, tunnel.ID, creator, msg.InitialDeposit); err != nil {
			return nil, err
		}
	}

	return &types.MsgCreateTunnelResponse{
		TunnelID: tunnel.ID,
	}, nil
}

// UpdateRoute updates the route details based on the route type, allowing certain arguments to be updated.
func (k msgServer) UpdateRoute(
	goCtx context.Context,
	msg *types.MsgUpdateRoute,
) (*types.MsgUpdateRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := k.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	if tunnel.Route.TypeUrl != msg.Route.TypeUrl {
		return nil, types.ErrInvalidRoute.Wrap("cannot change route type")
	}

	route, err := msg.GetRouteValue()
	if err != nil {
		return nil, err
	}

	switch r := route.(type) {
	case *types.IBCRoute:
		_, found := k.channelKeeper.GetChannel(ctx, PortIDForTunnel(msg.TunnelID), r.ChannelID)
		if !found {
			return nil, types.ErrInvalidChannelID
		}
		tunnel.Route = msg.Route

	default:
		return nil, types.ErrInvalidRoute.Wrap("cannot update route on this route type")
	}

	k.Keeper.SetTunnel(ctx, tunnel)

	return &types.MsgUpdateRouteResponse{}, nil
}

// UpdateSignalsAndInterval update signals and interval for a tunnel.
func (k msgServer) UpdateSignalsAndInterval(
	goCtx context.Context,
	msg *types.MsgUpdateSignalsAndInterval,
) (*types.MsgUpdateSignalsAndIntervalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.Keeper.GetParams(ctx)

	// validate signal infos and interval
	if err := types.ValidateSignalDeviations(msg.SignalDeviations, params.MaxSignals, params.MaxDeviationBPS, params.MinDeviationBPS); err != nil {
		return nil, err
	}

	// validate interval
	if err := types.ValidateInterval(msg.Interval, params.MaxInterval, params.MinInterval); err != nil {
		return nil, err
	}

	tunnel, err := k.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	err = k.Keeper.UpdateSignalsAndInterval(ctx, msg.TunnelID, msg.SignalDeviations, msg.Interval)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateSignalsAndIntervalResponse{}, nil
}

// WithdrawFeePayerFunds withdraws the fee payer's funds from the tunnel to the creator.
func (k msgServer) WithdrawFeePayerFunds(
	goCtx context.Context,
	msg *types.MsgWithdrawFeePayerFunds,
) (*types.MsgWithdrawFeePayerFundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := k.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	feePayer, err := sdk.AccAddressFromBech32(tunnel.FeePayer)
	if err != nil {
		return nil, err
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	// send coins from the fee payer to the creator
	if err := k.Keeper.bankKeeper.SendCoins(
		ctx,
		feePayer,
		creator,
		msg.Amount,
	); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawFeePayerFundsResponse{}, nil
}

// Activate activates a tunnel.
func (k msgServer) Activate(
	goCtx context.Context,
	msg *types.MsgActivate,
) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := k.Keeper.GetTunnel(ctx, msg.TunnelID)
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

	if err := k.Keeper.ActivateTunnel(ctx, msg.TunnelID); err != nil {
		return nil, err
	}

	return &types.MsgActivateResponse{}, nil
}

// Deactivate deactivates a tunnel.
func (k msgServer) Deactivate(
	goCtx context.Context,
	msg *types.MsgDeactivate,
) (*types.MsgDeactivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := k.Keeper.GetTunnel(ctx, msg.TunnelID)
	if err != nil {
		return nil, err
	}

	if msg.Creator != tunnel.Creator {
		return nil, types.ErrInvalidTunnelCreator.Wrapf("creator %s, tunnelID %d", msg.Creator, msg.TunnelID)
	}

	if !tunnel.IsActive {
		return nil, types.ErrAlreadyInactive.Wrapf("tunnelID %d", msg.TunnelID)
	}

	if err := k.Keeper.DeactivateTunnel(ctx, msg.TunnelID); err != nil {
		return nil, err
	}

	return &types.MsgDeactivateResponse{}, nil
}

// TriggerTunnel manually triggers a tunnel.
func (k msgServer) TriggerTunnel(
	goCtx context.Context,
	msg *types.MsgTriggerTunnel,
) (*types.MsgTriggerTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tunnel, err := k.Keeper.GetTunnel(ctx, msg.TunnelID)
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

	route, err := tunnel.GetRouteValue()
	if err != nil {
		return nil, err
	}

	// Check if the route is ready for receiving a new packet.
	if ok := k.IsRouteReady(ctx, route, tunnel.ID); !ok {
		return nil, types.ErrRouteNotReady.Wrapf("tunnelID %d", msg.TunnelID)
	}

	// Check if the fee payer of the tunnel has enough fund to create a packet.
	feePayer := sdk.MustAccAddressFromBech32(tunnel.FeePayer)
	ok, err := k.Keeper.HasEnoughFundToCreatePacket(ctx, route, feePayer)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, types.ErrInsufficientFund.Wrapf("tunnelID %d", msg.TunnelID)
	}

	signalIDs := tunnel.GetSignalIDs()
	prices := k.Keeper.feedsKeeper.GetPrices(ctx, signalIDs)

	// create a new packet
	packet, err := k.Keeper.CreatePacket(ctx, tunnel.ID, prices)
	if err != nil {
		return nil, err
	}

	// send packet
	if err := k.Keeper.SendPacket(ctx, packet); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to create packet for tunnel %d", tunnel.ID)
	}

	latestPrices, err := k.Keeper.GetLatestPrices(ctx, tunnel.ID)
	if err != nil {
		return nil, err
	}

	// update latest price info.
	latestPrices.LastInterval = ctx.BlockTime().Unix()
	latestPrices.UpdatePrices(packet.Prices)
	k.Keeper.SetLatestPrices(ctx, latestPrices)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTriggerTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", packet.TunnelID)),
		sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", packet.Sequence)),
	))

	return &types.MsgTriggerTunnelResponse{}, nil
}

// DepositToTunnel adds deposit to the tunnel.
func (k msgServer) DepositToTunnel(
	goCtx context.Context,
	msg *types.MsgDepositToTunnel,
) (*types.MsgDepositToTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.DepositToTunnel(ctx, msg.TunnelID, depositor, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgDepositToTunnelResponse{}, nil
}

// WithdrawFromTunnel withdraws deposit from the tunnel.
func (k msgServer) WithdrawFromTunnel(
	goCtx context.Context,
	msg *types.MsgWithdrawFromTunnel,
) (*types.MsgWithdrawFromTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	withdrawer, err := sdk.AccAddressFromBech32(msg.Withdrawer)
	if err != nil {
		return nil, err
	}

	// Withdraw the deposit from the tunnel
	if err := k.Keeper.WithdrawFromTunnel(ctx, msg.TunnelID, msg.Amount, withdrawer); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawFromTunnelResponse{}, nil
}

// UpdateParams updates the module params.
func (k msgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.Keeper.GetAuthority() != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			k.Keeper.GetAuthority(),
			msg.Authority,
		)
	}

	if err := k.Keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, msg.Params.String()),
	))

	return &types.MsgUpdateParamsResponse{}, nil
}
