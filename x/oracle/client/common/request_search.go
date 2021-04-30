package common

import (
	"fmt"
	"github.com/GeoDB-Limited/odin-core/hooks/common"
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"sort"
)

func queryLatestRequest(clientCtx client.Context, requestSearchRequest *oracletypes.QueryRequestSearchRequest) (oracletypes.RequestID, error) {
	bin := clientCtx.JSONMarshaler.MustMarshalJSON(requestSearchRequest)
	bz, _, err := clientCtx.QueryWithData(fmt.Sprintf("%s/%s/1", common.AppHook, oracletypes.QueryLatestRequest), bin)
	if err != nil {
		return 0, err
	}
	var containerIDs oracletypes.QueryRequestIDs
	err = clientCtx.JSONMarshaler.UnmarshalJSON(bz, &containerIDs)
	if err != nil {
		return 0, err
	}
	if len(containerIDs.RequestIds) == 0 {
		return 0, nil
	}
	if len(containerIDs.RequestIds) > 1 {
		// NEVER EXPECT TO HIT.
		panic("multi request limit=1")
	}

	return oracletypes.RequestID(containerIDs.RequestIds[0]), nil
}

func queryRequest(route string, clientCtx client.Context, rid oracletypes.RequestID) (oracletypes.QueryRequestResponse, int64, error) {
	bz, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%d", route, oracletypes.QueryRequests, rid))
	if err != nil {
		return oracletypes.QueryRequestResponse{}, 0, err
	}

	var queryResult commontypes.QueryResult
	if err := clientCtx.LegacyAmino.UnmarshalJSON(bz, &queryResult); err != nil {
		return oracletypes.QueryRequestResponse{}, 0, err
	}

	var result oracletypes.QueryRequestResponse
	if err := clientCtx.LegacyAmino.UnmarshalJSON(queryResult.Result, &result); err != nil {
		return oracletypes.QueryRequestResponse{}, 0, err
	}

	return result, height, nil
}

func QuerySearchLatestRequest(
	route string, clientCtx client.Context, requestSearchRequest *oracletypes.QueryRequestSearchRequest,
) (*oracletypes.QueryRequestSearchResponse, int64, error) {
	id, err := queryLatestRequest(clientCtx, requestSearchRequest)
	if err != nil {
		return nil, 0, err
	}

	if id == 0 {
		return nil, 0, nil
	}

	req, h, err := queryRequest(route, clientCtx, id)
	return oracletypes.NewQueryRequestSearchResponse(req), h, err
}

func queryMultiRequest(clientCtx client.Context, requestSearchParams *oracletypes.QueryRequestSearchRequest, limit int) (*oracletypes.QueryRequestIDs, error) {
	bin := clientCtx.JSONMarshaler.MustMarshalJSON(requestSearchParams)

	bz, _, err := clientCtx.QueryWithData(fmt.Sprintf("%s/%s/%d", common.AppHook, oracletypes.QueryLatestRequest, limit), bin)
	if err != nil {
		return nil, err
	}
	var containerIDs oracletypes.QueryRequestIDs
	err = clientCtx.JSONMarshaler.UnmarshalJSON(bz, &containerIDs)
	if err != nil {
		return nil, err
	}
	return &containerIDs, nil
}

func queryRequests(
	route string, clientCtx client.Context, containerIDs *oracletypes.QueryRequestIDs,
) ([]oracletypes.QueryRequestResponse, int64, error) {
	type queryResult struct {
		result oracletypes.QueryRequestResponse
		err    error
		height int64
	}
	queryResultsChan := make(chan queryResult, len(containerIDs.RequestIds))
	for _, rid := range containerIDs.RequestIds {
		go func(rid int64) {
			out, h, err := queryRequest(route, clientCtx, oracletypes.RequestID(rid))
			if err != nil {
				queryResultsChan <- queryResult{err: err}
				return
			}
			queryResultsChan <- queryResult{result: out, height: h}
		}(rid)
	}
	requests := make([]oracletypes.QueryRequestResponse, 0)
	height := int64(0)
	for idx := 0; idx < len(containerIDs.RequestIds); idx++ {
		select {
		case req := <-queryResultsChan:
			if req.err != nil {
				return nil, 0, req.err
			}
			if req.result.Request.ResponsePacketData.Result != nil {
				requests = append(requests, req.result)
				if req.height > height {
					height = req.height
				}
			}
		}
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Request.ResponsePacketData.ResolveTime > requests[j].Request.ResponsePacketData.ResolveTime
	})

	return requests, height, nil
}

func QueryMultiSearchLatestRequest(
	route string, clientCtx client.Context, requestSearchParams *oracletypes.QueryRequestSearchRequest, limit int,
) ([]oracletypes.QueryRequestResponse, int64, error) {
	requestIDs, err := queryMultiRequest(clientCtx, requestSearchParams, limit)
	if err != nil {
		return nil, 0, err
	}
	queryRequestResults, h, err := queryRequests(route, clientCtx, requestIDs)
	if err != nil {
		return nil, 0, err
	}
	if len(queryRequestResults) == 0 {
		return nil, 0, nil
	}
	if len(queryRequestResults) > limit {
		queryRequestResults = queryRequestResults[:limit]
	}
	return queryRequestResults, h, nil
}
