package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) types.QueryServer {
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
			case types.TUNNEL_STATUS_FILTER_UNSPECIFIED:
				return t, nil
			case types.TUNNEL_STATUS_FILTER_ACTIVE:
				if t.IsActive {
					return t, nil
				}
			case types.TUNNEL_STATUS_FILTER_INACTIVE:
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

// Deposit queries a deposit by its tunnel ID and depositor address.
func (q queryServer) Deposit(c context.Context, req *types.QueryDepositRequest) (*types.QueryDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, err
	}

	deposit, found := q.k.GetDeposit(ctx, req.TunnelId, depositor)
	if !found {
		return nil, types.ErrDepositNotFound
	}

	return &types.QueryDepositResponse{Deposit: deposit}, nil
}

// Deposits queries all deposits of the tunnel.
func (q queryServer) Deposits(
	c context.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var deposits []*types.Deposit
	depositStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.DepositsStoreKey(req.TunnelId))

	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key []byte, value []byte) error {
		var deposit types.Deposit
		if err := q.k.cdc.Unmarshal(value, &deposit); err != nil {
			return err
		}

		deposits = append(deposits, &deposit)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}

// Packets queries all packets of the module.
func (q queryServer) Packets(c context.Context, req *types.QueryPacketsRequest) (*types.QueryPacketsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	store := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.TunnelPacketsStoreKey(req.TunnelId))
	filteredPackets, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		store,
		req.Pagination,
		func(key []byte, p *types.Packet) (*types.Packet, error) {
			return p, nil
		}, func() *types.Packet {
			return &types.Packet{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPacketsResponse{Packets: filteredPackets, Pagination: pageRes}, nil
}

// Packet queries a packet by its ID.
func (q queryServer) Packet(c context.Context, req *types.QueryPacketRequest) (*types.QueryPacketResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	packet, err := q.k.GetPacket(ctx, req.TunnelId, req.Nonce)
	if err != nil {
		return nil, err
	}
	return &types.QueryPacketResponse{Packet: &packet}, nil
}

// Params queries all params of the module.
func (q queryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}
