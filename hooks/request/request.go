package request

import (
	"encoding/hex"
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

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// Hook inherits from Band app hook to save latest request into SQL database.
type Hook struct {
	cdc          codec.Codec
	oracleKeeper keeper.Keeper
	db           *gorm.DB
	dbMaxRecords int
	trans        *gorm.DB
}

var _ common.Hook = &Hook{}

// NewHook creates a request hook instance that will be added in Band App.
func NewHook(cdc codec.Codec, oracleKeeper keeper.Keeper, connStr string, numRecords int) *Hook {
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
			valAddr, err := sdk.ValAddressFromBech32(validator)
			if err != nil {
				ctx.Logger().
					Error("Unable to parse validator address got from EventTypeReport for request search", "error", err)
				continue
			}
			report, err := h.oracleKeeper.GetReport(ctx, reqID, valAddr)
			if err != nil {
				ctx.Logger().Error("Unable to get report for request search", "error", err)
				continue
			}
			// Collect reports which are submitted AFTER the request successfully resolved
			if !report.InBeforeResolve {
				res := h.oracleKeeper.MustGetResult(ctx, reqID)
				if res.ResolveStatus == types.RESOLVE_STATUS_SUCCESS {
					reports[reqID] = append(reports[reqID], report)
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
				request, err := h.oracleKeeper.GetRequest(ctx, reqID)
				if err != nil {
					ctx.Logger().Error("Unable to get request for request search", "reqID", reqID, "err", err)
					continue
				}
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

func (h *Hook) RequestSearch(req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, bool, error) {
	calldata, err := hex.DecodeString(req.Calldata)
	if err != nil {
		return nil, true, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unable to parse calldata: %s", err)
	}

	// Query oracle requests from database
	oracleReq, err := h.getLatestRequest(
		types.OracleScriptID(req.OracleScriptId),
		calldata,
		req.AskCount,
		req.MinCount,
	)

	// check query results
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, true, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "request not found")
		}
		return nil, true, sdkerrors.Wrap(sdkerrors.ErrLogic, "unable to query latest request from database")
	}

	queryResponse := oracleReq.QueryRequestResponse()
	return &types.QueryRequestSearchResponse{Request: &queryResponse}, true, nil
}

func (h *Hook) RequestPrice(req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, bool, error) {
	return nil, false, nil
}

// BeforeCommit specify actions need to do before commit block (app.Hook interface).
func (h *Hook) BeforeCommit() {
	err := h.trans.Commit()
	if err != nil {
		h.trans.Rollback()
	}
}
