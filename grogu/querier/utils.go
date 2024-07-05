package querier

import (
	"context"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type QueryFunction[I, O any] func(ctx context.Context, in *I, opts ...grpc.CallOption) (*O, error)

type responseWithBlockHeight[T any] struct {
	response    *T
	blockHeight int64
}

func getResponseWithBlockHeight[I, O any](
	f QueryFunction[I, O],
	in *I,
	opts ...grpc.CallOption,
) (*responseWithBlockHeight[O], error) {
	var header metadata.MD

	opts = append(opts, grpc.Header(&header))
	resp, err := f(context.Background(), in, opts...)
	if err != nil {
		return nil, err
	}

	blockHeightArr := header.Get(sdk.GRPCBlockHeightHeader)
	if len(blockHeightArr) == 0 {
		return nil, fmt.Errorf("block height not found in header")
	}

	blockHeight, err := strconv.ParseInt(blockHeightArr[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block height with error: %v", err)
	}

	return &responseWithBlockHeight[O]{resp, blockHeight}, nil
}

func getResponseWithBlockHeightToChannel[I, O any](
	resultCh chan *responseWithBlockHeight[O],
	errorCh chan error,
	f QueryFunction[I, O],
	in *I,
	opts ...grpc.CallOption,
) {
	resp, err := getResponseWithBlockHeight(f, in, opts...)
	if err != nil {
		errorCh <- err
		return
	}

	resultCh <- resp
}

func getMaxBlockHeightResponse[I, O any](
	fs []QueryFunction[I, O],
	in *I,
	opts ...grpc.CallOption,
) (*O, error) {
	resultCh := make(chan *responseWithBlockHeight[O], len(fs))
	errorCh := make(chan error, len(fs))

	for _, f := range fs {
		go getResponseWithBlockHeightToChannel(resultCh, errorCh, f, in, opts...)
	}

	var resp *O
	var maxBlockHeight int64
	var err error
	for range fs {
		select {
		case r := <-resultCh:
			if r.blockHeight <= maxBlockHeight {
				continue
			}

			resp = r.response
			maxBlockHeight = r.blockHeight
		case err = <-errorCh:
			continue
		}
	}

	if resp != nil {
		return resp, nil
	}

	return nil, err
}
