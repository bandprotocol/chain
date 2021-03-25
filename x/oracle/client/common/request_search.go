package common

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"sort"

// 	"github.com/cosmos/cosmos-sdk/client"

// 	"github.com/bandprotocol/chain/x/oracle/types"
// )

// func queryLatestRequest(clientCtx client.Context, oid, calldata, askCount, minCount string) (types.RequestID, error) {
// 	bz, _, err := clientCtx.Query(fmt.Sprintf("band/latest_request/%s/%s/%s/%s/1", oid, calldata, askCount, minCount))
// 	if err != nil {
// 		return 0, err
// 	}
// 	var reqIDs []types.RequestID
// 	err = clientCtx.LegacyAmino.UnmarshalBinaryBare(bz, &reqIDs)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if len(reqIDs) == 0 {
// 		return 0, errors.New("request with specified specification not found")
// 	}
// 	if len(reqIDs) > 1 {
// 		// NEVER EXPECT TO HIT.
// 		panic("multi request limit=1")
// 	}

// 	return reqIDs[0], nil
// }

// func queryRequest(clientCtx client.Context, rid types.RequestID) (types.QueryRequestResult, int64, error) {
// 	bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%d", types.QueryRequests, rid))
// 	if err != nil {
// 		return types.QueryRequestResult{}, 0, err
// 	}
// 	var result types.QueryResult
// 	if err := json.Unmarshal(bz, &result); err != nil {
// 		return types.QueryRequestResult{}, 0, err
// 	}
// 	var reqResult types.QueryRequestResult
// 	clientCtx.LegacyAmino.MustUnmarshalJSON(result.Result, &reqResult)
// 	return reqResult, height, nil
// }

// func QuerySearchLatestRequest(
// 	clientCtx client.Context, oid, calldata, askCount, minCount string,
// ) ([]byte, int64, error) {
// 	id, err := queryLatestRequest(clientCtx, oid, calldata, askCount, minCount)
// 	if err != nil {
// 		bz, err := types.QueryNotFound(clientCtx.LegacyAmino, "request with specified specification not found")
// 		return bz, 0, err
// 	}
// 	out, h, err := queryRequest(clientCtx, id)
// 	bz, err := types.QueryOK(clientCtx.LegacyAmino, out)
// 	return bz, h, err
// }

// func queryMultitRequest(clientCtx client.Context, oid, calldata, askCount, minCount string, limit int) ([]types.RequestID, error) {
// 	bz, _, err := clientCtx.Query(fmt.Sprintf("band/latest_request/%s/%s/%s/%s/%d", oid, calldata, askCount, minCount, limit))
// 	if err != nil {
// 		return nil, err
// 	}
// 	var reqIDs []types.RequestID
// 	err = clientCtx.LegacyAmino.UnmarshalBinaryBare(bz, &reqIDs)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return reqIDs, nil
// }

// func queryRequests(
// 	clientCtx client.Context, requestIDs []types.RequestID,
// ) ([]types.QueryRequestResult, int64, error) {
// 	type queryResult struct {
// 		result types.QueryRequestResult
// 		err    error
// 		height int64
// 	}
// 	queryResultsChan := make(chan queryResult, len(requestIDs))
// 	for _, rid := range requestIDs {
// 		go func(rid types.RequestID) {
// 			out, h, err := queryRequest(clientCtx, rid)
// 			if err != nil {
// 				queryResultsChan <- queryResult{err: err}
// 				return
// 			}
// 			queryResultsChan <- queryResult{result: out, height: h}
// 		}(rid)
// 	}
// 	requests := make([]types.QueryRequestResult, 0)
// 	height := int64(0)
// 	for idx := 0; idx < len(requestIDs); idx++ {
// 		select {
// 		case req := <-queryResultsChan:
// 			if req.err != nil {
// 				return nil, 0, req.err
// 			}
// 			if req.result.Result != nil {
// 				requests = append(requests, req.result)
// 				if req.height > height {
// 					height = req.height
// 				}
// 			}
// 		}
// 	}

// 	sort.Slice(requests, func(i, j int) bool {
// 		return requests[i].Result.ResponsePacketData.ResolveTime > requests[j].Result.ResponsePacketData.ResolveTime
// 	})

// 	return requests, height, nil
// }

// func QueryMultiSearchLatestRequest(
// 	clientCtx client.Context, oid, calldata, askCount, minCount string, limit int,
// ) ([]byte, int64, error) {
// 	requestIDs, err := queryMultitRequest(clientCtx, oid, calldata, askCount, minCount, limit)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	queryRequestResults, h, err := queryRequests(clientCtx, requestIDs)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	if len(queryRequestResults) == 0 {
// 		bz, err := types.QueryNotFound(clientCtx.LegacyAmino, "request with specified specification not found")
// 		return bz, 0, err
// 	}
// 	if len(queryRequestResults) > limit {
// 		queryRequestResults = queryRequestResults[:limit]
// 	}
// 	bz, err := types.QueryOK(clientCtx.LegacyAmino, queryRequestResults)
// 	return bz, h, err
// }
