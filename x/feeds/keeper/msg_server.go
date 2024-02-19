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

	// check whether delegator has enough delegation for signals
	sumPower := sumPower(req.Signals)
	sumDelegation := ms.Keeper.stakingKeeper.GetDelegatorBonded(ctx, delegator).Uint64()
	if sumPower > sumDelegation {
		return nil, types.ErrNotEnoughDelegation
	}

	// delete previous signal, decrease symbol power by the previous signals
	symbolToIntervalDiff := make(map[string]int64)
	prevSignals := ms.Keeper.GetDelegatorSignals(ctx, delegator)
	for _, prevSignal := range prevSignals {
		symbol, err := ms.Keeper.GetSymbol(ctx, prevSignal.Symbol)
		if err != nil {
			return nil, err
		}
		// before changing in symbol, delete the SymbolByPower index
		ms.Keeper.DeleteSymbolByPowerIndex(ctx, symbol)

		symbol.Power -= prevSignal.Power
		prevInterval := symbol.Interval
		symbol.Interval = calculateInterval(int64(symbol.Power), ms.Keeper.GetParams(ctx))
		ms.Keeper.SetSymbol(ctx, symbol)

		// setting SymbolByPowerIndex every time setting symbol
		ms.Keeper.SetSymbolByPowerIndex(ctx, symbol)

		intervalDiff := symbol.Interval - prevInterval
		if symbol.Interval-prevInterval != 0 {
			symbolToIntervalDiff[symbol.Symbol] = intervalDiff
		}
	}

	// increase symbol power by the new signals
	ms.Keeper.SetDelegatorSignals(ctx, delegator, types.Signals{Signals: req.Signals})
	for _, signal := range req.Signals {
		symbol, err := ms.Keeper.GetSymbol(ctx, signal.Symbol)
		if err != nil {
			symbol = types.Symbol{
				Symbol:                      signal.Symbol,
				Power:                       0,
				Interval:                    0,
				LastIntervalUpdateTimestamp: 0,
			}
		}
		// before changing in symbol, delete the SymbolByPower index
		ms.Keeper.DeleteSymbolByPowerIndex(ctx, symbol)

		symbol.Power += signal.Power
		prevInterval := symbol.Interval
		symbol.Interval = calculateInterval(int64(symbol.Power), ms.Keeper.GetParams(ctx))
		ms.Keeper.SetSymbol(ctx, symbol)

		// setting SymbolByPowerIndex every time setting symbol
		ms.Keeper.SetSymbolByPowerIndex(ctx, symbol)

		// if the sum interval differences is zero then the interval is not changed
		intervalDiff := (symbol.Interval - prevInterval) + symbolToIntervalDiff[symbol.Symbol]
		if intervalDiff == 0 {
			delete(symbolToIntervalDiff, symbol.Symbol)
		} else {
			symbolToIntervalDiff[symbol.Symbol] = intervalDiff
		}
	}

	// update interval timestamp for interval-changed symbols
	for symbolName := range symbolToIntervalDiff {
		symbol, err := ms.Keeper.GetSymbol(ctx, symbolName)
		if err != nil {
			return nil, err
		}
		// before changing in symbol, delete the SymbolByPower index
		ms.Keeper.DeleteSymbolByPowerIndex(ctx, symbol)

		symbol.LastIntervalUpdateTimestamp = ctx.BlockTime().Unix()
		ms.Keeper.SetSymbol(ctx, symbol)

		// setting SymbolByPowerIndex every time setting symbol
		ms.Keeper.SetSymbolByPowerIndex(ctx, symbol)
	}

	return &types.MsgSignalSymbolsResponse{}, nil
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
		return nil, types.ErrNotTopValidator
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
	transitionTime := ms.Keeper.GetParams(ctx).TransitionTime

	for _, price := range req.Prices {
		s, err := ms.Keeper.GetSymbol(ctx, price.Symbol)
		if err != nil {
			return nil, err
		}

		priceVal, err := ms.Keeper.GetPriceValidator(ctx, price.Symbol, val)
		if err == nil {
			if blockTime < priceVal.Timestamp+s.Interval-transitionTime {
				return nil, types.ErrPriceTooFast.Wrapf(
					"symbol: %s, old: %d, new: %d, interval: %d",
					price.Symbol,
					priceVal.Timestamp,
					blockTime,
					s.Interval,
				)
			}
		}

		priceVal = types.PriceValidator{
			Validator: req.Validator,
			Symbol:    price.Symbol,
			Price:     price.Price,
			Timestamp: blockTime,
		}

		_ = ms.Keeper.SetPriceValidator(ctx, priceVal)

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
