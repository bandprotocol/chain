package price

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/syndtr/goleveldb/leveldb"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/hooks/common"
	"github.com/bandprotocol/chain/pkg/obi"
	"github.com/bandprotocol/chain/x/oracle/keeper"
	"github.com/bandprotocol/chain/x/oracle/types"
)

// Hook uses levelDB to store the latest price of standard price reference.
type Hook struct {
	cdc          codec.Marshaler
	stdOs        map[types.OracleScriptID]bool
	oracleKeeper keeper.Keeper
	db           *leveldb.DB
}

// NewHook creates a price hook instance that will be added in Band App.
func NewHook(cdc codec.Marshaler, oracleKeeper keeper.Keeper, oids []types.OracleScriptID, priceDBDir string) *Hook {
	stdOs := make(map[types.OracleScriptID]bool)
	for _, oid := range oids {
		stdOs[oid] = true
	}
	db, err := leveldb.OpenFile(priceDBDir, nil)
	if err != nil {
		panic(err)
	}
	return &Hook{
		cdc:          cdc,
		stdOs:        stdOs,
		oracleKeeper: oracleKeeper,
		db:           db,
	}
}

// AfterInitChain specify actions need to do after chain initialization (app.Hook interface).
func (h *Hook) AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, res abci.ResponseInitChain) {
}

// AfterBeginBlock specify actions need to do after begin block period (app.Hook interface).
func (h *Hook) AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
}

// AfterDeliverTx specify actions need to do after transaction has been processed (app.Hook interface).
func (h *Hook) AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
}

// AfterEndBlock specify actions need to do after end block period (app.Hook interface).
func (h *Hook) AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) {
	for _, event := range res.Events {
		events := sdk.StringifyEvents([]abci.Event{event})
		evMap := common.ParseEvents(events)
		switch event.Type {
		case types.EventTypeResolve:
			reqID := types.RequestID(common.Atoi(evMap[types.EventTypeResolve+"."+types.AttributeKeyID][0]))
			result := h.oracleKeeper.MustGetResult(ctx, reqID)

			if result.ResolveStatus == types.RESOLVE_STATUS_SUCCESS {
				// Check that we need to store data to db
				if h.stdOs[result.OracleScriptID] {
					var input Input
					var output Output
					obi.MustDecode(result.Calldata, &input)
					obi.MustDecode(result.Result, &output)
					for idx, symbol := range input.Symbols {
						price := types.QueryRequestPriceResponse{
							Symbol:      symbol,
							Multiplier:  input.Multiplier,
							Rate:        output.Rates[idx],
							RequestID:   result.RequestID,
							ResolveTime: result.ResolveTime,
						}
						err := h.db.Put([]byte(fmt.Sprintf("%d,%d,%s", result.AskCount, result.MinCount, symbol)),
							h.cdc.MustMarshalBinaryBare(&price), nil)
						if err != nil {
							panic(err)
						}
					}
				}
			}
		}
	}
}

// CustomQuery catch the custom query that matches specific paths (app.Hook interface).
func (h *Hook) CustomQuery(path string) (res []byte, match bool, err error) {
	paths := strings.Split(path, "/")
	switch paths[0] {
	case "prices":
		if len(paths) < 4 {
			return nil, true, errors.New("no route for prices query specified")
		}
		symbol := paths[1]
		askCount := common.Atoui(paths[2])
		minCount := common.Atoui(paths[3])
		bz, err := h.db.Get([]byte(fmt.Sprintf("%d,%d,%s", askCount, minCount, symbol)), nil)
		if err != nil {
			return nil, true, fmt.Errorf(
				"cannot get price of %s with %d/%d counts with error: %s",
				symbol, minCount, askCount, err.Error(),
			)
		}
		return bz, true, nil
	default:
		return nil, false, nil
	}
}

// BeforeCommit specify actions need to do before commit block (app.Hook interface).
func (h *Hook) BeforeCommit() {}
