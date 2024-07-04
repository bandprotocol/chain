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

func (ms msgServer) CreateTunnel(
	goCtx context.Context,
	req *types.MsgCreateTunnel,
) (*types.MsgCreateTunnelResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// ms.Keeper.AddTunnel(ctx, types.Tunnel{
	// 	Route:    req.Route,
	// 	FeedType: req.FeedType,
	// })

	fmt.Printf("Msg Create Tunnel: %+v\n", req.Route.TypeUrl)
	fmt.Printf("Msg ROute: %+v\n", req.GetTunnelRoute())
	fmt.Printf("Msg ROute: %+v\n", req.Route.GetCachedValue().(types.Route))
	// switch req.Route.GetCachedValue().(type) {
	// case *types.TSSRoute:
	// 	// Validate TSSRoute
	// 	fmt.Printf("TSSRoute\n")
	// case *types.AxelarRoute:
	// 	// Validate AxelarRoute
	// 	fmt.Printf("AxelarRoute\n")
	// default:
	// 	return &types.MsgCreateTunnelResponse{}, sdkerrors.ErrUnknownRequest.Wrapf("unknown route type")
	// }

	return &types.MsgCreateTunnelResponse{}, nil
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
