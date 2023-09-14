package bandibctesting

import (
	"encoding/json"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/bandprotocol/chain/v2/testing/testapp"
)

type TestChain struct {
	*ibctesting.TestChain
}

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	ta, genesis := testapp.NewTestApp("BANDCHAIN", log.NewNopLogger())
	return ta, genesis
}

func (chain *TestChain) SetActiveValidators() error {
	for _, val := range chain.Vals.Validators {
		err := chain.GetBandApp().OracleKeeper.Activate(chain.GetContext(), sdk.ValAddress(val.Address))
		if err != nil {
			return err
		}
	}
	return nil
}

func (chain *TestChain) SendMoneyToValidators() error {
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 10000))

	for _, val := range chain.Vals.Validators {
		err := chain.GetBandApp().BankKeeper.SendCoins(
			chain.GetContext(),
			chain.SenderAccount.GetAddress(),
			sdk.AccAddress(val.Address),
			sendCoins,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetBandApp returns the current chain's app as an BandApp
func (chain *TestChain) GetBandApp() *testapp.TestingApp {
	app, ok := chain.App.(*testapp.TestingApp)
	if !ok {
		panic("not transfer app")
	}

	return app
}
