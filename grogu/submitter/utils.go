package submitter

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/grogu/querier"
)

func getTxResponse(
	txQuerier *querier.TxQuerier,
	txHash string,
	timeout time.Duration,
	pollInterval time.Duration,
) (*sdk.TxResponse, error) {
	var resp *sdk.TxResponse
	var err error

	for start := time.Now(); time.Since(start) < timeout; {
		time.Sleep(pollInterval)
		resp, err = txQuerier.QueryTx(txHash)
		if err != nil {
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("timeout exceeded with error: %v", err)
}

func unpackAccount(account *client.Account, resp *auth.QueryAccountResponse) error {
	registry := band.MakeEncodingConfig().InterfaceRegistry
	err := registry.UnpackAny(resp.Account, account)
	if err != nil {
		return fmt.Errorf("failed to unpack account with error: %v", err)
	}

	return nil
}
