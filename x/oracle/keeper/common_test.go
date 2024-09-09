package keeper_test

import (
	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	chainID            = "BANDCHAIN"
	basicName          = "BASIC_NAME"
	basicDesc          = "BASIC_DESCRIPTION"
	basicSchema        = "BASIC_SCHEMA"
	basicSourceCodeURL = "BASIC_SOURCE_CODE_URL"
	basicFilename      = "BASIC_FILENAME"
	basicClientID      = "BASIC_CLIENT_ID"
)

var (
	owner    = sdk.AccAddress([]byte("owner_______________"))
	treasury = sdk.AccAddress([]byte("treasury____________"))
	alice    = sdk.AccAddress([]byte("alice_______________"))
	bob      = sdk.AccAddress([]byte("bob_________________"))

	PKS = simtestutil.CreateTestPubKeys(3)

	valConsPk0 = PKS[0]
	valConsPk1 = PKS[1]
	valConsPk2 = PKS[2]

	validators = []ValidatorWithValAddress{
		createValidator(PKS[0], math.NewInt(100)),
		createValidator(PKS[1], math.NewInt(70)),
		createValidator(PKS[2], math.NewInt(30)),
	}

	basicCalldata = []byte("BASIC_CALLDATA")
	basicReport   = []byte("BASIC_REPORT")
	basicResult   = []byte("BASIC_RESULT")

	emptyCoins        = sdk.Coins(nil)
	coinsZero         = sdk.NewCoins()
	coins10uband      = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	coins20uband      = sdk.NewCoins(sdk.NewInt64Coin("uband", 20))
	coins1000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
)

type ValidatorWithValAddress struct {
	Validator stakingtypes.Validator
	Address   sdk.ValAddress
}

func createValidator(pk cryptotypes.PubKey, stake math.Int) ValidatorWithValAddress {
	valConsAddr := sdk.GetConsAddress(pk)
	val, _ := stakingtypes.NewValidator(
		sdk.ValAddress(valConsAddr).String(),
		pk,
		stakingtypes.Description{Moniker: "TestValidator"},
	)
	val.Tokens = stake
	val.DelegatorShares = math.LegacyNewDecFromInt(val.Tokens)
	return ValidatorWithValAddress{Validator: val, Address: sdk.ValAddress(valConsAddr)}
}
