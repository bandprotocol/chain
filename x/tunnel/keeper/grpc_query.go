package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k Keeper }

func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{k: k}
}

// Tunnels queries all tunnels.
func (q queryServer) Tunnels(c context.Context, req *types.QueryTunnelsRequest) (*types.QueryTunnelsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	store := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.TunnelStoreKeyPrefix)
	filteredTunnels, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		store,
		req.Pagination,
		func(key []byte, t *types.Tunnel) (*types.Tunnel, error) {
			// Filter by status
			switch req.IsActive {
			case types.TUNNEL_STATUS_UNSPECIFIED:
			case types.TUNNEL_STATUS_ACTIVE:
				if t.IsActive {
					return t, nil
				}
			case types.TUNNEL_STATUS_INACTIVE:
				if !t.IsActive {
					return t, nil
				}
			}

			return nil, nil
		}, func() *types.Tunnel {
			return &types.Tunnel{}
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryTunnelsResponse{Tunnels: filteredTunnels, Pagination: pageRes}, nil
}

// Tunnel queries a tunnel by its ID.
func (q queryServer) Tunnel(c context.Context, req *types.QueryTunnelRequest) (*types.QueryTunnelResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	tunnel, err := q.k.GetTunnel(ctx, req.TunnelId)
	if err != nil {
		return nil, err
	}

	return &types.QueryTunnelResponse{Tunnel: tunnel}, nil
}

// Params queries all params of the module.
func (q queryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}
