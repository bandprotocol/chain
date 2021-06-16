package request

import (
	"errors"
	"strings"

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
	dbMaxRecords int
	trans        *gorm.DB
}

var _ band.Hook = &Hook{}

// NewHook creates a request hook instance that will be added in Band App.
func NewHook(cdc codec.Marshaler, oracleKeeper keeper.Keeper, connStr string, numRecords int) *Hook {
	dbConnStr := strings.SplitN(connStr, ":", 2)
	for i := range dbConnStr {
		dbConnStr[i] = strings.TrimSpace(dbConnStr[i])
	}

	return &Hook{
		cdc:          cdc,
		oracleKeeper: oracleKeeper,
		db:           initDb(dbConnStr[0], dbConnStr[1]),
		dbMaxRecords: numRecords,
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
	reports := make(map[types.RequestID][]types.Report)
	for _, event := range res.Events {
		events := sdk.StringifyEvents([]abci.Event{event})
		evMap := common.ParseEvents(events)

		switch event.Type {
		case types.EventTypeReport:
			reqID := types.RequestID(common.Atoi(evMap[types.EventTypeReport+"."+types.AttributeKeyID][0]))
			validator := evMap[types.EventTypeReport+"."+types.AttributeKeyValidator][0]
			iter := h.oracleKeeper.GetReportIterator(ctx, reqID)
			defer iter.Close()

			// Check all reports related to given request ID
			for ; iter.Valid(); iter.Next() {
				var rep types.Report
				h.cdc.MustUnmarshalBinaryBare(iter.Value(), &rep)

				// Collect reports which are submitted after the request successfully resolved
				if !rep.InBeforeResolve && rep.Validator == validator {
					res := h.oracleKeeper.MustGetResult(ctx, reqID)
					if res.ResolveStatus == types.RESOLVE_STATUS_SUCCESS {
						reports[reqID] = append(reports[reqID], rep)
					}
				}
			}
		}
	}

	// Add collected reports to database
	if len(reports) > 0 {
		h.insertReports(reports)
	}
}

// AfterEndBlock specify actions need to do after end block period (app.Hook interface).
func (h *Hook) AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) {
	var requests []types.QueryRequestResponse
	for _, event := range res.Events {
		events := sdk.StringifyEvents([]abci.Event{event})
		evMap := common.ParseEvents(events)
		switch event.Type {
		case types.EventTypeResolve:
			reqID := types.RequestID(common.Atoi(evMap[types.EventTypeResolve+"."+types.AttributeKeyID][0]))

			// Collect resolved successful requests
			result := h.oracleKeeper.MustGetResult(ctx, reqID)
			if result.ResolveStatus == types.RESOLVE_STATUS_SUCCESS {
				request := h.oracleKeeper.MustGetRequest(ctx, reqID)
				reports := h.oracleKeeper.GetReports(ctx, reqID)
				requests = append(requests, types.QueryRequestResponse{
					Request: &request,
					Reports: reports,
					Result:  &result,
				})
			}
		}
	}

	// Add collected requsts to database
	if len(requests) > 0 {
		h.insertRequests(requests)
		for _, request := range requests {
			h.removeOldRecords(request)
		}
	}
}

// ApplyQuery catch the custom query that matches specific paths (app.Hook interface).
func (h *Hook) ApplyQuery(req abci.RequestQuery) (res abci.ResponseQuery, stop bool) {
	switch req.Path {
	case "/oracle.v1.Query/LatestRequest":
		var queryReq types.QueryLatestRequestRequest
		if err := h.cdc.UnmarshalBinaryBare(req.Data, &queryReq); err != nil {
			return sdkerrors.QueryResult(sdkerrors.Wrap(err, "unable to parse request data")), true
		}

		// Query oracle requests from database
		oracleReq, err := h.getLatestRequest(
			types.OracleScriptID(queryReq.OracleScriptId),
			queryReq.Calldata,
			queryReq.AskCount,
			queryReq.MinCount,
		)

		// If no result from database, use default value of qReqRes, which is nil
		var qReqRes *types.QueryRequestResponse
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return sdkerrors.QueryResult(sdkerrors.Wrap(err, "unable to query latest request from database")), true
		} else if err == nil {
			result := oracleReq.QueryRequestResponse()
			qReqRes = &result
		}

		finalResult := types.QueryLatestRequestResponse{
			Request: qReqRes,
		}

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
