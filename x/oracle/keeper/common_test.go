package keeper_test

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/pkg/filecache"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
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

	reporterPrivKey = secp256k1.GenPrivKey()
	reporterPubKey  = reporterPrivKey.PubKey()
	reporterAddr    = sdk.AccAddress(reporterPubKey.Address())

	basicCalldata                = []byte("BASIC_CALLDATA")
	basicReport                  = []byte("BASIC_REPORT")
	basicResult                  = []byte("BASIC_RESULT")
	testDefaultPrepareGas uint64 = 40000
	testDefaultExecuteGas uint64 = 300000

	emptyCoins          = sdk.Coins(nil)
	coinsZero           = sdk.NewCoins()
	coins10uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	coins20uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 20))
	coins1000000uband   = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	coins100000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 100000000))
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

func addSimpleDataSourceAndOracleScript(ctx sdk.Context, k keeper.Keeper, dir string) {
	// Add data source
	for i := 1; i <= 3; i++ {
		idxStr := fmt.Sprintf("%d", i)
		k.SetDataSource(
			ctx,
			types.DataSourceID(i),
			types.NewDataSource(
				owner,
				"name"+idxStr,
				"desc"+idxStr,
				"filename"+idxStr,
				coins1000000uband,
				treasury,
			),
		)
	}
	fc := filecache.New(dir)
	// Add wasm_1_simple
	fileName1 := fc.AddFile(testdata.Compile(testdata.Wasm1))
	k.SetOracleScript(
		ctx,
		types.OracleScriptID(1),
		types.NewOracleScript(owner, "test os", "testing oracle script", fileName1, "schema", "url"),
	)

	// Add wasm_4_complex
	fileName4 := fc.AddFile(testdata.Compile(testdata.Wasm4))
	k.SetOracleScript(
		ctx,
		types.OracleScriptID(4),
		types.NewOracleScript(owner, "test os4", "testing oracle script complex", fileName4, "schema", "url"),
	)
}

func defaultRequest() types.Request {
	return types.NewRequest(
		1,
		basicCalldata,
		[]sdk.ValAddress{validators[0].Address, validators[1].Address},
		2,
		1,
		bandtesting.ParseTime(0),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100000000uband,
	)
}
