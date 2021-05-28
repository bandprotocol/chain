package request

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gorm.io/gorm"

	// DB driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	abci "github.com/tendermint/tendermint/abci/types"

	band "github.com/bandprotocol/chain/app"
	"github.com/bandprotocol/chain/hooks/common"
	"github.com/bandprotocol/chain/x/oracle/keeper"
	"github.com/bandprotocol/chain/x/oracle/types"
)

// Hook inherits from Band app hook to save latest request into SQL database.
type Hook struct {
	cdc          codec.Marshaler
	oracleKeeper keeper.Keeper
	db           *gorm.DB
	trans        *gorm.DB
	baseApp      *baseapp.BaseApp
}

var _ band.Hook = &Hook{}

// NewHook creates a request hook instance that will be added in Band App.
func NewHook(cdc codec.Marshaler, oracleKeeper keeper.Keeper, connStr string, baseApp *baseapp.BaseApp) *Hook {
	dbConnStr := strings.SplitN(connStr, ":", 2)
	for i := range dbConnStr {
		dbConnStr[i] = strings.TrimSpace(dbConnStr[i])
	}

	return &Hook{
		cdc:          cdc,
		oracleKeeper: oracleKeeper,
		db:           initDb(dbConnStr[0], dbConnStr[1]),
		baseApp:      baseApp,
	}
}

// AfterInitChain specify actions need to do after chain initialization (app.Hook interface).
func (h *Hook) AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, res abci.ResponseInitChain) {
}

// AfterBeginBlock specify actions need to do after begin block period (app.Hook interface).
func (h *Hook) AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	trans := h.db.Begin()
	h.trans = trans
}

// AfterDeliverTx specify actions need to do after transaction has been processed (app.Hook interface).
func (h *Hook) AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
	for _, event := range res.Events {
		events := sdk.StringifyEvents([]abci.Event{event})
		evMap := common.ParseEvents(events)

		switch event.Type {
		case types.EventTypeReport:
			reqID := types.RequestID(common.Atoi(evMap[types.EventTypeReport+"."+types.AttributeKeyID][0]))
			validator := evMap[types.EventTypeReport+"."+types.AttributeKeyValidator][0]
			iter := h.oracleKeeper.GetReportIterator(ctx, reqID)
			defer iter.Close()

			for ; iter.Valid(); iter.Next() {
				var rep types.Report
				h.cdc.MustUnmarshalBinaryBare(iter.Value(), &rep)

				if !rep.InBeforeResolve && rep.Validator == validator {
					h.addReport(reqID, rep)
				}
			}
		}
	}
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
				request := h.oracleKeeper.MustGetRequest(ctx, reqID)
				reports := h.oracleKeeper.GetReports(ctx, reqID)
				h.insertRequest(request, reports, result)
			}
		}
	}
}

// ApplyQuery catch the custom query that matches specific paths (app.Hook interface).
func (h *Hook) ApplyQuery(req abci.RequestQuery) (res abci.ResponseQuery, stop bool) {
	switch req.Path {
	case "/oracle.v1.Query/RequestSearch":
		var request types.QueryRequestSearchRequest
		if err := h.cdc.UnmarshalBinaryBare(req.Data, &request); err != nil {
			return sdkerrors.QueryResult(sdkerrors.Wrap(err, "unable to parse request data")), true
		}

		requests, err := h.getMultiRequests(
			types.OracleScriptID(request.OracleScriptId),
			request.Calldata,
			request.AskCount,
			request.MinCount,
			request.Limit,
		)
		if err != nil {
			return sdkerrors.QueryResult(sdkerrors.Wrap(err, "unable to query multiple requests from database")), true
		}

		finalResult := requests.QueryRequestSearchResponse()

		bz, err := h.cdc.MarshalBinaryBare(&finalResult)
		if err != nil {
			return common.QueryResultError(err), true
		}
		return common.QueryResultSuccess(bz, req.Height), true
	default:
		return abci.ResponseQuery{}, false
	}
}

// BeforeCommit specify actions need to do before commit block (app.Hook interface).
func (h *Hook) BeforeCommit() {
	err := h.trans.Commit()
	if err != nil {
		h.trans.Rollback()
	}
}
