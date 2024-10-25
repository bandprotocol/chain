package submitter

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	band "github.com/bandprotocol/chain/v3/app"
)

func unpackAccount(
	registry codectypes.InterfaceRegistry,
	account *client.Account,
	resp *auth.QueryAccountResponse,
) error {
	err := registry.UnpackAny(resp.Account, account)
	if err != nil {
		return fmt.Errorf("failed to unpack account with error: %v", err)
	}

	return nil
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", ".band")
	if err != nil {
		dir = band.DefaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}
