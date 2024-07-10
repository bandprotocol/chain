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

	signalPriceInfos := make([]types.SignalPriceInfo, len(req.SignalInfos))
	for _, signalInfo := range req.SignalInfos {
		signalPriceInfos = append(signalPriceInfos, types.SignalPriceInfo{
			SignalID:     signalInfo.SignalID,
			DeviationBPS: signalInfo.DeviationBPS,
			Interval:     signalInfo.Interval,
		})
	}

	tunnelID, err := ms.Keeper.AddTunnel(ctx, types.Tunnel{
		Route:            req.Route,
		FeedType:         req.FeedType,
		SignalPriceInfos: signalPriceInfos,
		IsActive:         true, // TODO: set to false by default
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
		sdk.NewAttribute(types.AttributeRoute, tunnel.Route.String()),
		sdk.NewAttribute(types.AttributeFeedType, tunnel.FeedType.String()),
		sdk.NewAttribute(types.AttributeFeePayer, tunnel.FeePayer),
		sdk.NewAttribute(types.AttributeIsActive, fmt.Sprintf("%t", tunnel.IsActive)),
		sdk.NewAttribute(types.AttributeKeyCreatedAt, tunnel.CreatedAt.String()),
		sdk.NewAttribute(types.AttributeCreator, req.Creator),
	)
	for _, signalInfo := range req.SignalInfos {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeSignalPriceInfos, signalInfo.String()),
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
		sdk.NewAttribute(types.AttributeIsActive, fmt.Sprintf("%t", true)),
	))

	return &types.MsgActivateTunnelResponse{}, nil
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
