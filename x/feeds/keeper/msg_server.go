package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/feeds MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

func (ms msgServer) UpdateSymbols(
	goCtx context.Context,
	req *types.MsgUpdateSymbols,
) (*types.MsgUpdateSymbolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin := ms.GetParams(ctx).Admin
	if admin != req.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			req.Admin,
		)
	}

	// TODO
	// - reject duplicated symbol

	for _, symbol := range req.Symbols {
		time := ctx.BlockTime().Unix()

		// if the symbols already existing, we won't update timestamp.
		s, err := ms.Keeper.GetSymbol(ctx, symbol.Symbol)
		if err == nil {
			time = s.Timestamp
		}

		ms.Keeper.SetSymbol(ctx, types.Symbol{
			Symbol:    symbol.Symbol,
			Interval:  symbol.Interval,
			Timestamp: time,
		})

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUpdateSymbol,
				sdk.NewAttribute(types.AttributeKeySymbol, symbol.Symbol),
				sdk.NewAttribute(types.AttributeKeyInterval, fmt.Sprintf("%d", symbol.Interval)),
			),
		)
	}

	return &types.MsgUpdateSymbolsResponse{}, nil
}

func (ms msgServer) RemoveSymbols(
	goCtx context.Context,
	req *types.MsgRemoveSymbols,
) (*types.MsgRemoveSymbolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin := ms.GetParams(ctx).Admin
	if admin != req.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			req.Admin,
		)
	}

	// TODO
	// - reject duplicated symbol

	for _, symbol := range req.Symbols {
		ms.Keeper.DeleteSymbol(ctx, symbol)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRemoveSymbol,
				sdk.NewAttribute(types.AttributeKeySymbol, symbol),
			),
		)
	}

	return &types.MsgRemoveSymbolsResponse{}, nil
}

func (ms msgServer) SubmitPrices(
	goCtx context.Context,
	req *types.MsgSubmitPrices,
) (*types.MsgSubmitPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO:
	// - check validator is in top 100 ?

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	status := ms.Keeper.oracleKeeper.GetValidatorStatus(ctx, val)
	if !status.IsActive {
		return nil, types.ErrOracleStatusNotActive.Wrapf("val: %s", val.String())
	}

	if req.Timestamp > int64(ctx.BlockTime().Unix())+60 {
		return nil, types.ErrInvalidTimestamp.Wrapf(
			"block_time: %d, timestamp: %d",
			ctx.BlockTime().Unix(),
			req.Timestamp,
		)
	}

	for _, price := range req.Prices {
		if _, err := ms.Keeper.GetSymbol(ctx, price.Symbol); err != nil {
			return nil, err
		}

		priceVal, err := ms.Keeper.GetPriceValidator(ctx, price.Symbol, val)
		if err == nil {
			if req.Timestamp < priceVal.Timestamp {
				return nil, types.ErrTimestampOlder.Wrapf(
					"symbol: %s, current: %d, new: %d",
					price.Symbol,
					priceVal.Timestamp,
					req.Timestamp,
				)
			}
		}

		priceVal = types.PriceValidator{
			Validator: req.Validator,
			Symbol:    price.Symbol,
			Price:     price.Price,
			Timestamp: ctx.BlockTime().Unix(),
		}

		ms.Keeper.SetPriceValidator(ctx, priceVal)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitPrice,
				sdk.NewAttribute(types.AttributeKeyValidator, priceVal.Validator),
				sdk.NewAttribute(types.AttributeKeySymbol, priceVal.Symbol),
				sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", priceVal.Price)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", priceVal.Timestamp)),
			),
		)
	}

	return &types.MsgSubmitPricesResponse{}, nil
}

func (ms msgServer) UpdateOffChain(
	goCtx context.Context,
	req *types.MsgUpdateOffChain,
) (*types.MsgUpdateOffChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin := ms.GetParams(ctx).Admin
	if admin != req.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			req.Admin,
		)
	}

	if err := ms.SetOffChain(ctx, req.OffChain); err != nil {
		return nil, err
	}

	return &types.MsgUpdateOffChainResponse{}, nil
}

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

	return &types.MsgUpdateParamsResponse{}, nil
}
