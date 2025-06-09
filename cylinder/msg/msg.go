package msg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RequestType represents the type of the request.
type RequestType int

const (
	RequestTypeCreateGroupRound1 RequestType = iota
	RequestTypeCreateGroupRound2
	RequestTypeCreateGroupConfirm
	RequestTypeCreateGroupComplain
	RequestTypeUpdateDE
	RequestTypeSubmitSignature
)

// Request is the struct for sending a request to the sender worker.
type Request struct {
	ReqType RequestType
	ID      uint64
	Msg     sdk.Msg
	Retry   uint64
}

// NewRequest creates a new request object.
func NewRequest(reqType RequestType, id uint64, msg sdk.Msg) Request {
	return Request{
		ReqType: reqType,
		ID:      id,
		Msg:     msg,
		Retry:   0,
	}
}

// IncreaseRetry increases the retry count of the request and returns a new request.
func (req Request) IncreaseRetry() Request {
	return Request{
		ReqType: req.ReqType,
		ID:      req.ID,
		Msg:     req.Msg,
		Retry:   req.Retry + 1,
	}
}

// Response is the struct for identifying the result of the request.
type Response struct {
	Request Request
	Success bool
	TxHash  string
	Err     error
}

// NewResponse creates a new response object.
func NewResponse(req Request, success bool, txHash string, err error) Response {
	return Response{
		Request: req,
		Success: success,
		TxHash:  txHash,
		Err:     err,
	}
}

// ResponseReceiver is the struct for receiving the response from the sender worker.
type ResponseReceiver struct {
	ReqType    RequestType
	ResponseCh chan Response
}

// NewResponseReceiver creates a new response receiver for a request type.
func NewResponseReceiver(reqType RequestType) ResponseReceiver {
	return ResponseReceiver{
		ReqType:    reqType,
		ResponseCh: make(chan Response, 1000),
	}
}
