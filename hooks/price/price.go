package price

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

// Hook uses levelDB to store the latest price of standard price reference.
type Hook struct {
	cdc             codec.Codec
	stdOs           map[types.OracleScriptID]bool
	oracleKeeper    keeper.Keeper
	db              *leveldb.DB
	defaultAskCount uint64
	defaultMinCount uint64
}

var _ common.Hook = &Hook{}

// NewHook creates a price hook instance that will be added in Band App.
func NewHook(
	cdc codec.Codec,
	oracleKeeper keeper.Keeper,
	oids []types.OracleScriptID,
	priceDBDir string,
	defaultAskCount uint64,
	defaultMinCount uint64,
) *Hook {
	stdOs := make(map[types.OracleScriptID]bool)
	for _, oid := range oids {
		stdOs[oid] = true
	}
	db, err := leveldb.OpenFile(priceDBDir, nil)
	if err != nil {
		panic(err)
	}
	return &Hook{
		cdc:             cdc,
		stdOs:           stdOs,
		oracleKeeper:    oracleKeeper,
		db:              db,
		defaultAskCount: defaultAskCount,
		defaultMinCount: defaultMinCount,
	}
}

// AfterInitChain specify actions need to do after chain initialization (app.Hook interface).
func (h *Hook) AfterInitChain(ctx sdk.Context, req *abci.RequestInitChain, res *abci.ResponseInitChain) {
}

// AfterBeginBlock specify actions need to do after begin block period (app.Hook interface).
func (h *Hook) AfterBeginBlock(ctx sdk.Context, req *abci.RequestFinalizeBlock, events []abci.Event) {
}

// AfterDeliverTx specify actions need to do after transaction has been processed (app.Hook interface).
func (h *Hook) AfterDeliverTx(ctx sdk.Context, tx sdk.Tx, res *abci.ExecTxResult) {
}

// AfterEndBlock specify actions need to do after end block period (app.Hook interface).
func (h *Hook) AfterEndBlock(ctx sdk.Context, events []abci.Event) {
	for _, event := range events {
		events := sdk.StringifyEvents([]abci.Event{event})
		evMap := common.ParseEvents(events)
		switch event.Type {
		case types.EventTypeResolve:
			reqID := types.RequestID(common.Atoi(evMap[types.EventTypeResolve+"."+types.AttributeKeyID][0]))
			result := h.oracleKeeper.MustGetResult(ctx, reqID)

			if result.ResolveStatus == types.RESOLVE_STATUS_SUCCESS {
				// check whether this result should be stored to database
				if h.stdOs[result.OracleScriptID] {
					commonOutput := MustDecodeResult(result.Calldata, result.Result)
					for idx, symbol := range commonOutput.Symbols {
						price := types.PriceResult{
							Symbol:      symbol,
							Multiplier:  commonOutput.Multiplier,
							Px:          commonOutput.Rates[idx],
							RequestID:   result.RequestID,
							ResolveTime: result.ResolveTime,
						}
						err := h.db.Put([]byte(fmt.Sprintf("%d,%d,%s", result.AskCount, result.MinCount, symbol)),
							h.cdc.MustMarshal(&price), nil)
						if err != nil {
							panic(err)
						}
					}
				}
			}
		default:
			// No action needed for other cases yet
		}
	}
}

func (h *Hook) RequestSearch(req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, bool, error) {
	return nil, false, nil
}

func (h *Hook) RequestPrice(req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, bool, error) {
	var response types.QueryRequestPriceResponse
	for _, symbol := range req.Symbols {
		var priceResult types.PriceResult

		if req.AskCount == 0 && req.MinCount == 0 {
			req.AskCount = h.defaultAskCount
			req.MinCount = h.defaultMinCount
		}
		bz, err := h.db.Get([]byte(fmt.Sprintf("%d,%d,%s", req.AskCount, req.MinCount, symbol)), nil)
		if err != nil {
			if errors.Is(err, leveldb.ErrNotFound) {
				return nil, true, sdkerrors.ErrKeyNotFound.Wrapf(
					"price not found for %s with %d/%d counts",
					symbol,
					req.AskCount,
					req.MinCount,
				)
			}
			return nil, true, sdkerrors.ErrLogic.Wrapf(
				"unable to get price of %s with %d/%d counts",
				symbol,
				req.AskCount,
				req.MinCount,
			)
		}
		h.cdc.MustUnmarshal(bz, &priceResult)
		response.PriceResults = append(response.PriceResults, &priceResult)
	}
	return &response, true, nil
}

// BeforeCommit specify actions need to do before commit block (app.Hook interface).
func (h *Hook) BeforeCommit() {}
