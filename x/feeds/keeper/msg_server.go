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

func (ms msgServer) SignalSymbols(
	goCtx context.Context,
	req *types.MsgSignalSymbols,
) (*types.MsgSignalSymbolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegator, err := sdk.AccAddressFromBech32(req.Delegator)
	if err != nil {
		return nil, err
	}

	sumPower := sumPower(req.Signal.Symbols)
	sumDelegation := ms.Keeper.GetDelegatorDelegationsSum(ctx, delegator)
	if sumPower > sumDelegation {
		return nil, types.ErrNotEnoughDelegation
	}
	oldSignal, found := ms.Keeper.GetDelegatorSignal(ctx, delegator)
	if found {
		for _, symbolWithPower := range oldSignal.Symbols {
			power := ms.Keeper.GetSymbolPower(ctx, symbolWithPower.Symbol)
			power = power - symbolWithPower.Power
			ms.Keeper.SetSymbolPower(ctx, symbolWithPower.Symbol, power)
		}
	}

	ms.Keeper.SetDelegatorSignal(ctx, delegator, *req.Signal)
	for _, symbolWithPower := range req.Signal.Symbols {
		power := ms.Keeper.GetSymbolPower(ctx, symbolWithPower.Symbol)
		power = power + symbolWithPower.Power
		ms.Keeper.SetSymbolPower(ctx, symbolWithPower.Symbol, power)
	}

	return &types.MsgSignalSymbolsResponse{}, nil
}

func sumPower(symbolsWithPower []types.SymbolWithPower) (sum uint64) {
	for _, symbol := range symbolsWithPower {
		sum = sum + symbol.Power
	}
	return
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

	for _, symbol := range req.Symbols {
		time := ctx.BlockTime().Unix()

		// if the symbols already existing, we won't update timestamp.
		s, err := ms.Keeper.GetSymbol(ctx, symbol.Symbol)
		if err == nil {
			time = s.Timestamp
		}

		ms.Keeper.SetSymbol(ctx, types.Symbol{
			Symbol:      symbol.Symbol,
			MinInterval: symbol.MinInterval,
			MaxInterval: symbol.MaxInterval,
			Timestamp:   time,
		})

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUpdateSymbol,
				sdk.NewAttribute(types.AttributeKeySymbol, symbol.Symbol),
				sdk.NewAttribute(types.AttributeKeyMinInterval, fmt.Sprintf("%d", symbol.MinInterval)),
				sdk.NewAttribute(types.AttributeKeyMaxInterval, fmt.Sprintf("%d", symbol.MaxInterval)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", time)),
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

	for _, symbol := range req.Symbols {
		if _, err := ms.Keeper.GetSymbol(ctx, symbol); err != nil {
			return nil, err
		}

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
	blockTime := ctx.BlockTime().Unix()

	// check if it's in top bonded validators.
	vals := ms.stakingKeeper.GetBondedValidatorsByPower(ctx)
	isInTop := false
	for _, val := range vals {
		if req.Validator == val.GetOperator().String() {
			isInTop = true
			break
		}
	}
	if !isInTop {
		return nil, types.ErrInvalidTimestamp
	}

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	status := ms.Keeper.oracleKeeper.GetValidatorStatus(ctx, val)
	if !status.IsActive {
		return nil, types.ErrOracleStatusNotActive.Wrapf("val: %s", val.String())
	}

	if types.AbsInt64(req.Timestamp-blockTime) > ms.Keeper.GetParams(ctx).AllowDiffTime {
		return nil, types.ErrInvalidTimestamp.Wrapf(
			"block_time: %d, timestamp: %d",
			blockTime,
			req.Timestamp,
		)
	}

	for _, price := range req.Prices {
		s, err := ms.Keeper.GetSymbol(ctx, price.Symbol)
		if err != nil {
			return nil, err
		}

		priceVal, err := ms.Keeper.GetPriceValidator(ctx, price.Symbol, val)
		if err == nil {
			if blockTime < priceVal.Timestamp+s.MinInterval {
				return nil, types.ErrPriceTooFast.Wrapf(
					"symbol: %s, old: %d, new: %d, min_interval: %d",
					price.Symbol,
					priceVal.Timestamp,
					blockTime,
					s.MinInterval,
				)
			}
		}

		priceVal = types.PriceValidator{
			Validator: req.Validator,
			Symbol:    price.Symbol,
			Price:     price.Price,
			Timestamp: blockTime,
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

func (ms msgServer) UpdatePriceService(
	goCtx context.Context,
	req *types.MsgUpdatePriceService,
) (*types.MsgUpdatePriceServiceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin := ms.GetParams(ctx).Admin
	if admin != req.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			req.Admin,
		)
	}

	if err := ms.SetPriceService(ctx, req.PriceService); err != nil {
		return nil, err
	}

	return &types.MsgUpdatePriceServiceResponse{}, nil
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
