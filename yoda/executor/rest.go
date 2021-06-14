package executor

import (
	"encoding/base64"
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"net/url"
	"time"

	"github.com/levigross/grequests"
)

type RestExec struct {
	url     string
	timeout time.Duration
}

func NewRestExec(url string, timeout time.Duration) *RestExec {
	return &RestExec{url: url, timeout: timeout}
}

type externalExecutionResponse struct {
	ReturnCode uint32 `json:"returncode"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Version    string `json:"version"`
}

func (e *RestExec) Exec(code []byte, arg string, env interface{}) (ExecResult, error) {
	executable := base64.StdEncoding.EncodeToString(code)
	resp, err := grequests.Post(
		e.url,
		&grequests.RequestOptions{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			JSON: map[string]interface{}{
				"executable": executable,
				"calldata":   arg,
				"timeout":    e.timeout.Milliseconds(),
				"env":        env,
			},
			RequestTimeout: e.timeout,
		},
	)
	if err != nil {
		urlErr, ok := err.(*url.Error)
		if !ok || !urlErr.Timeout() {
			return ExecResult{}, err
		}
		// Return timeout code
		return ExecResult{Output: []byte{}, Code: 111}, nil
	}

	if resp.Ok != true {
		return ExecResult{}, sdkerrors.Wrap(errors.ErrNotOkResponse, "execution failed")
	}

	r := externalExecutionResponse{}
	if err := resp.JSON(&r); err != nil {
		return ExecResult{}, sdkerrors.Wrap(err, "failed to parse the execution response")
	}

	if r.ReturnCode == 0 {
		return ExecResult{Output: []byte(r.Stdout), Code: 0, Version: r.Version}, nil
	} else {
		return ExecResult{Output: []byte(r.Stderr), Code: r.ReturnCode, Version: r.Version}, nil
	}
}
