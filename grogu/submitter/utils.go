package submitter

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	band "github.com/bandprotocol/chain/v3/app"
)

func unpackAccount(account *client.Account, resp *auth.QueryAccountResponse) error {
	initAppOptions := viper.New()
	tempDir := tempDir()
	initAppOptions.Set(flags.FlagHome, tempDir)
	tempApplication := band.NewBandApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tempDir,
		initAppOptions,
		100,
	)
	registry := tempApplication.InterfaceRegistry()
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
