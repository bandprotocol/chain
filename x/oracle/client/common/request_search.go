package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GeoDB-Limited/odin-core/hooks/common"
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"sort"
)

// TODO: remove ???
func queryLatestRequest(clientCtx client.Context, requestSearchParams oracletypes.QueryRequestSearchParams) (oracletypes.RequestID, error) {
	bin := clientCtx.LegacyAmino.MustMarshalJSON(requestSearchParams)
	bz, _, err := clientCtx.QueryWithData(fmt.Sprintf("%s/%s/%d/1", common.AppHook, oracletypes.QueryLatestRequest, requestSearchParams.OracleScriptID), bin)
	if err != nil {
		return 0, err
	}
	var reqIDs []oracletypes.RequestID
	err = clientCtx.LegacyAmino.UnmarshalBinaryBare(bz, &reqIDs)
	if err != nil {
		return 0, err
	}
	if len(reqIDs) == 0 {
		return 0, errors.New("request with specified specification not found")
	}
	if len(reqIDs) > 1 {
		// NEVER EXPECT TO HIT.
		panic("multi request limit=1")
	}

	return reqIDs[0], nil
}

func queryRequest(route string, clientCtx client.Context, rid oracletypes.RequestID) (oracletypes.QueryRequestResult, int64, error) {
	bz, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%d", route, oracletypes.QueryRequests, rid))
	if err != nil {
		return oracletypes.QueryRequestResult{}, 0, err
	}

	var result commontypes.QueryResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return oracletypes.QueryRequestResult{}, 0, err
	}

	var reqResult oracletypes.QueryRequestResult
	clientCtx.LegacyAmino.MustUnmarshalJSON(result.Result, &reqResult)
	return reqResult, height, nil
}

// TODO: remove ???
func QuerySearchLatestRequest(
	route string, clientCtx client.Context, requestSearchParams oracletypes.QueryRequestSearchParams,
) ([]byte, int64, error) {
	id, err := queryLatestRequest(clientCtx, requestSearchParams)
	if err != nil {
		bz, err := commontypes.QueryNotFound(clientCtx.LegacyAmino, "request with specified specification not found")
		return bz, 0, err
	}
	out, h, err := queryRequest(route, clientCtx, id)
	bz, err := commontypes.QueryOK(clientCtx.LegacyAmino, out)
	return bz, h, err
}

func queryMultiRequest(clientCtx client.Context, requestSearchParams oracletypes.QueryRequestSearchParams, limit int) ([]oracletypes.RequestID, error) {
	bin := clientCtx.LegacyAmino.MustMarshalJSON(requestSearchParams)

	bz, _, err := clientCtx.QueryWithData(fmt.Sprintf("%s/%s/%d/%d", common.AppHook, oracletypes.QueryLatestRequest, requestSearchParams.OracleScriptID, limit), bin)
	if err != nil {
		return nil, err
	}
	var reqIDs []oracletypes.RequestID
	err = clientCtx.LegacyAmino.UnmarshalBinaryBare(bz, &reqIDs)
	if err != nil {
		return nil, err
	}
	return reqIDs, nil
}

func queryRequests(
	route string, clientCtx client.Context, requestIDs []oracletypes.RequestID,
) ([]oracletypes.QueryRequestResult, int64, error) {
	type queryResult struct {
		result oracletypes.QueryRequestResult
		err    error
		height int64
	}
	queryResultsChan := make(chan queryResult, len(requestIDs))
	for _, rid := range requestIDs {
		go func(rid oracletypes.RequestID) {
			out, h, err := queryRequest(route, clientCtx, rid)
			if err != nil {
				queryResultsChan <- queryResult{err: err}
				return
			}
			queryResultsChan <- queryResult{result: out, height: h}
		}(rid)
	}
	requests := make([]oracletypes.QueryRequestResult, 0)
	height := int64(0)
	for idx := 0; idx < len(requestIDs); idx++ {
		select {
		case req := <-queryResultsChan:
			if req.err != nil {
				return nil, 0, req.err
			}
			if req.result.Result != nil {
				requests = append(requests, req.result)
				if req.height > height {
					height = req.height
				}
			}
		}
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Result.ResolveTime > requests[j].Result.ResolveTime
	})

	return requests, height, nil
}

func QueryMultiSearchLatestRequest(
	route string, clientCtx client.Context, requestSearchParams oracletypes.QueryRequestSearchParams, limit int,
) ([]byte, int64, error) {
	requestIDs, err := queryMultiRequest(clientCtx, requestSearchParams, limit)
	if err != nil {
		return nil, 0, err
	}
	queryRequestResults, h, err := queryRequests(route, clientCtx, requestIDs)
	if err != nil {
		return nil, 0, err
	}
	if len(queryRequestResults) == 0 {
		bz, err := commontypes.QueryNotFound(clientCtx.LegacyAmino, "request with specified specification not found")
		return bz, 0, err
	}
	if len(queryRequestResults) > limit {
		queryRequestResults = queryRequestResults[:limit]
	}
	bz, err := commontypes.QueryOK(clientCtx.LegacyAmino, queryRequestResults)
	return bz, h, err
}
