package executor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createDefaultServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		ret := externalExecutionResponse{
			Returncode: 0,
			Stdout:     "BEEB",
			Stderr:     "Stderr",
		}
		_ = json.NewEncoder(res).Encode(ret)
	}))
}

func createResponseNotOkScenarioServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(500)
	}))
}

func createCannotDecodeJSONScenarioServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		_, _ = res.Write([]byte("invalid bytes"))
	}))
}

func createExecuteFailScenarioServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		ret := externalExecutionResponse{
			Returncode: 1,
			Stdout:     "BEEB",
			Stderr:     "Stderr",
		}
		_ = json.NewEncoder(res).Encode(ret)
	}))
}

func TestExecuteSuccess(t *testing.T) {
	testServer := createDefaultServer()
	defer func() { testServer.Close() }()

	executor := NewRestExec(testServer.URL, 1*time.Second)
	res, err := executor.Exec([]byte("executable"), "calldata", nil)
	require.NoError(t, err)
	require.Equal(t, uint32(0), res.Code)
	require.Equal(t, []byte("BEEB"), res.Output)
}

func TestExecuteBadUrlFail(t *testing.T) {
	testServer := createDefaultServer()
	defer func() { testServer.Close() }()

	executor := NewRestExec("www.beeb.com", 1*time.Second) // bad url
	_, err := executor.Exec([]byte("executable"), "calldata", nil)
	require.Error(t, err)
}

func TestExecuteDecodeStructFail(t *testing.T) {
	testServer := createCannotDecodeJSONScenarioServer()
	defer func() { testServer.Close() }()

	executor := NewRestExec(testServer.URL, 1*time.Second)
	_, err := executor.Exec([]byte("executable"), "calldata", nil)
	require.Error(t, err)
}

func TestExecuteResponseNotOk(t *testing.T) {
	testServer := createResponseNotOkScenarioServer()
	defer func() { testServer.Close() }()

	executor := NewRestExec(testServer.URL, 1*time.Second)
	_, err := executor.Exec([]byte("executable"), "calldata", nil)
	require.ErrorIs(t, err, ErrRestNotOk)
}

func TestExecuteFail(t *testing.T) {
	testServer := createExecuteFailScenarioServer()
	defer func() { testServer.Close() }()

	executor := NewRestExec(testServer.URL, 1*time.Second)
	res, err := executor.Exec([]byte("executable"), "calldata", nil)
	require.NoError(t, err)
	require.Equal(t, uint32(1), res.Code)
	require.Equal(t, []byte("Stderr"), res.Output)
}
