package submitter

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	band "github.com/bandprotocol/chain/v2/app"
)

func unpackAccount(account *client.Account, resp *auth.QueryAccountResponse) error {
	registry := band.MakeEncodingConfig().InterfaceRegistry
	err := registry.UnpackAny(resp.Account, account)
	if err != nil {
		return fmt.Errorf("failed to unpack account with error: %v", err)
	}

	return nil
}
