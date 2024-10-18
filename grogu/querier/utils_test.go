package querier

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	sdk "github.com/cosmos/cosmos-sdk/types/grpc"
)

// Mock Request and Response types
type MockRequest struct{}

type MockResponse struct {
	Value string
}

// Mock gRPC function generator
func generateMockFuncWithBlockHeight(blockHeight string) QueryFunction[MockRequest, MockResponse] {
	return func(ctx context.Context, in *MockRequest, opts ...grpc.CallOption) (*MockResponse, error) {
		var header *metadata.MD
		for _, opt := range opts {
			if hOpt, ok := opt.(grpc.HeaderCallOption); ok {
				header = hOpt.HeaderAddr
				*header = metadata.Pairs(sdk.GRPCBlockHeightHeader, blockHeight)
			}
		}
		return &MockResponse{Value: "mock response"}, nil
	}
}

func mockFuncMissingHeader(ctx context.Context, in *MockRequest, opts ...grpc.CallOption) (*MockResponse, error) {
	return &MockResponse{Value: "mock response"}, nil
}

func TestGetResponseWithBlockHeight(t *testing.T) {
	in := &MockRequest{}
	opts := []grpc.CallOption{}

	resp, err := getResponseWithBlockHeight(generateMockFuncWithBlockHeight("15"), in, opts...)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int64(15), resp.blockHeight)
	require.Equal(t, "mock response", resp.response.Value)
}

func TestGetResponseWithBlockHeight_MissingHeader(t *testing.T) {
	in := &MockRequest{}
	opts := []grpc.CallOption{}

	resp, err := getResponseWithBlockHeight(mockFuncMissingHeader, in, opts...)
	require.Error(t, err)
	require.Contains(t, err.Error(), "block height not found in header")
	require.Nil(t, resp)
}

func TestGetResponseWithBlockHeight_InvalidHeader(t *testing.T) {
	in := &MockRequest{}
	opts := []grpc.CallOption{}

	resp, err := getResponseWithBlockHeight(generateMockFuncWithBlockHeight("invalid"), in, opts...)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse block height")
	require.Nil(t, resp)
}

func TestGetMaxBlockHeightResponse(t *testing.T) {
	fs := []QueryFunction[MockRequest, MockResponse]{
		generateMockFuncWithBlockHeight("1"),
		generateMockFuncWithBlockHeight("12"),
		generateMockFuncWithBlockHeight("15"),
	}

	in := &MockRequest{}
	maxBlockHeight := new(atomic.Int64)
	maxBlockHeight.Store(10)
	opts := []grpc.CallOption{}

	resp, err := getMaxBlockHeightResponse(fs, in, maxBlockHeight, opts...)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int64(15), maxBlockHeight.Load())
	require.Equal(t, "mock response", resp.Value)
}

func TestGetMaxBlockHeightResponse_LowerThanCurrentMax(t *testing.T) {
	fs := []QueryFunction[MockRequest, MockResponse]{
		generateMockFuncWithBlockHeight("19"),
	}

	in := &MockRequest{}
	maxBlockHeight := new(atomic.Int64)
	maxBlockHeight.Store(20)
	opts := []grpc.CallOption{}

	resp, err := getMaxBlockHeightResponse(fs, in, maxBlockHeight, opts...)
	require.Error(t, err)
	require.Contains(t, err.Error(), "block height is lower than latest max block height")
	require.Nil(t, resp)
}

func TestGetMaxBlockHeightResponse_AllFail(t *testing.T) {
	fs := []QueryFunction[MockRequest, MockResponse]{
		func(ctx context.Context, in *MockRequest, opts ...grpc.CallOption) (*MockResponse, error) {
			return nil, fmt.Errorf("failed")
		},
		func(ctx context.Context, in *MockRequest, opts ...grpc.CallOption) (*MockResponse, error) {
			return nil, fmt.Errorf("failed")
		},
	}

	in := &MockRequest{}
	maxBlockHeight := new(atomic.Int64)
	maxBlockHeight.Store(10)
	opts := []grpc.CallOption{}

	_, err := getMaxBlockHeightResponse(fs, in, maxBlockHeight, opts...)
	require.Error(t, err)
}
