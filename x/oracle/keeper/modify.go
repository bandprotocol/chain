package keeper

import (
	"github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// modify returns new value if it is not `DoNotModify`. Returns old value otherwise
func modify(oldVal string, newVal string) string {
	if newVal == types.DoNotModify {
		return oldVal
	}
	return newVal
}

func modifyAddress(oldVal string, newVal string) string {
	if newVal == sdk.AccAddress(types.DoNotModifyBytes).String() {
		return oldVal
	}
	return newVal
}

func modifyCoins(oldVal sdk.Coins, newVal sdk.Coins) sdk.Coins {
	if newVal.IsEqual(types.DoNotModifyCoins) {
		return oldVal
	}
	return newVal
}
