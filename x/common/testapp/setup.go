package testapp

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	bandapp "github.com/GeoDB-Limited/odin-core/app"
	"github.com/GeoDB-Limited/odin-core/pkg/filecache"
	me "github.com/GeoDB-Limited/odin-core/x/oracle/keeper"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	owasm "github.com/bandprotocol/go-owasm/api"
)

// Account is a data structure to store key of test account.
type Account struct {
	PrivKey    crypto.PrivKey
	PubKey     crypto.PubKey
	Address    sdk.AccAddress
	ValAddress sdk.ValAddress
}

// nolint
var (
	Owner              Account
	Treasury           Account
	FeePayer           Account
	Peter              Account
	Alice              Account
	Bob                Account
	Carol              Account
	OraclePoolProvider Account
	FeePoolProvider    Account
	Validators         []Account
	DataSources        []oracletypes.DataSource
	OracleScripts      []oracletypes.OracleScript
	OwasmVM            *owasm.Vm
)

// nolint
var (
	EmptyCoins               = sdk.Coins(nil)
	Coin1geo                 = sdk.NewInt64Coin("geo", 1)
	Coin10odin               = sdk.NewInt64Coin("odin", 10)
	Coin100000000geo         = sdk.NewInt64Coin("geo", 100000000)
	Coins1000000odin         = sdk.NewCoins(sdk.NewInt64Coin("odin", 1000000))
	Coins99999999odin        = sdk.NewCoins(sdk.NewInt64Coin("odin", 99999999))
	Coin100000000odin        = sdk.NewInt64Coin("odin", 100000000)
	Coin10000000000odin      = sdk.NewInt64Coin("odin", 10000000000)
	Coins100000000odin       = sdk.NewCoins(Coin100000000odin)
	Coins10000000000odin     = sdk.NewCoins(Coin10000000000odin)
	DefaultDataProvidersPool = sdk.NewCoins(Coin100000000odin)
	DefaultCommunityPool     = sdk.NewCoins(Coin100000000geo, Coin100000000odin)
)

func init() {
	bandapp.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
	r := rand.New(rand.NewSource(time.Now().Unix()))
	Owner = createArbitraryAccount(r)
	Treasury = createArbitraryAccount(r)
	FeePayer = createArbitraryAccount(r)
	Peter = createArbitraryAccount(r)
	Alice = createArbitraryAccount(r)
	Bob = createArbitraryAccount(r)
	Carol = createArbitraryAccount(r)
	OraclePoolProvider = createArbitraryAccount(r)
	FeePoolProvider = createArbitraryAccount(r)
	for i := 0; i < 3; i++ {
		Validators = append(Validators, createArbitraryAccount(r))
	}
	owasmVM, err := owasm.NewVm(10)
	if err != nil {
		panic(err)
	}
	OwasmVM = owasmVM
}

func createArbitraryAccount(r *rand.Rand) Account {
	privkeySeed := make([]byte, 12)
	r.Read(privkeySeed)
	privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)
	return Account{
		PrivKey:    privKey,
		PubKey:     privKey.PubKey(),
		Address:    sdk.AccAddress(privKey.PubKey().Address()),
		ValAddress: sdk.ValAddress(privKey.PubKey().Address()),
	}
}

func getGenesisDataSources(homePath string) []oracletypes.DataSource {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	DataSources = []oracletypes.DataSource{{}} // 0th index should be ignored
	for idx := 0; idx < 5; idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile([]byte("code" + idxStr))
		DataSources = append(DataSources, oracletypes.NewDataSource(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, Coins1000000odin,
		))
	}
	return DataSources[1:]
}

func getGenesisOracleScripts(homePath string) []oracletypes.OracleScript {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	OracleScripts = []oracletypes.OracleScript{{}} // 0th index should be ignored
	wasms := [][]byte{
		Wasm1, Wasm2, Wasm3, Wasm4, Wasm56(10), Wasm56(10000000), Wasm78(10), Wasm78(2000), Wasm9,
	}
	for idx := 0; idx < len(wasms); idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile(compile(wasms[idx]))
		OracleScripts = append(OracleScripts, oracletypes.NewOracleScript(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, "schema"+idxStr, "url"+idxStr,
		))
	}
	return OracleScripts[1:]
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

// NewSimApp creates instance of our app using in test.
func NewSimApp(chainID string, logger log.Logger) *bandapp.BandApp {
	// Set HomeFlag to a temp folder for simulation run.
	dir, err := ioutil.TempDir("", "bandd")
	if err != nil {
		panic(err)
	}
	viper.Set(cli.HomeFlag, dir)

	db := dbm.NewMemDB()
	encCdc := bandapp.MakeEncodingConfig()
	app := bandapp.NewBandApp(logger, db, nil, true, map[int64]bool{}, dir, 0, encCdc, EmptyAppOptions{}, false, 0)

	genesis := bandapp.NewDefaultGenesisState()
	acc := []authtypes.GenesisAccount{
		&authtypes.BaseAccount{Address: Owner.Address.String()},
		&authtypes.BaseAccount{Address: FeePayer.Address.String()},
		&authtypes.BaseAccount{Address: Peter.Address.String()},
		&authtypes.BaseAccount{Address: Alice.Address.String()},
		&authtypes.BaseAccount{Address: Bob.Address.String()},
		&authtypes.BaseAccount{Address: Carol.Address.String()},
		&authtypes.BaseAccount{Address: Validators[0].Address.String()},
		&authtypes.BaseAccount{Address: Validators[1].Address.String()},
		&authtypes.BaseAccount{Address: Validators[2].Address.String()},
		&authtypes.BaseAccount{Address: OraclePoolProvider.Address.String()},
		&authtypes.BaseAccount{Address: FeePoolProvider.Address.String()},
	}
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), acc)
	genesis[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(Validators))
	bamt := []sdk.Int{Coins100000000odin[0].Amount, Coins1000000odin[0].Amount, Coins99999999odin[0].Amount}
	// bondAmt := sdk.NewInt(1000000)
	for idx, val := range Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			panic(err)
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			panic(err)
		}
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Unbonded,
			Tokens:            bamt[idx],
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(acc[4+idx].GetAddress(), val.Address.Bytes(), sdk.OneDec()))
	}
	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = "odin"
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesis[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	// Fund seed accounts and validators with 1000000odin and 100000000odin initially.
	balances := []banktypes.Balance{
		{
			Address: Owner.Address.String(),
			Coins:   Coins1000000odin,
		},
		{Address: FeePayer.Address.String(), Coins: Coins100000000odin},
		{Address: Peter.Address.String(), Coins: Coins1000000odin},
		{Address: Alice.Address.String(), Coins: Coins1000000odin.Add(Coin100000000geo)},
		{Address: Bob.Address.String(), Coins: Coins1000000odin},
		{Address: Carol.Address.String(), Coins: Coins1000000odin},
		{Address: Validators[0].Address.String(), Coins: Coins100000000odin},
		{Address: Validators[1].Address.String(), Coins: Coins100000000odin},
		{Address: Validators[2].Address.String(), Coins: Coins100000000odin},
		{Address: OraclePoolProvider.Address.String(), Coins: DefaultDataProvidersPool},
		{Address: FeePoolProvider.Address.String(), Coins: DefaultCommunityPool},
	}
	totalSupply := sdk.NewCoins()
	for idx := 0; idx < len(balances)-len(validators); idx++ {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(balances[idx].Coins...)
	}
	for idx := 0; idx < len(validators); idx++ {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(balances[idx+len(balances)-len(validators)].Coins.Add(sdk.NewCoin("odin", bamt[idx]))...)
	}

	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	// Add genesis data sources and oracle scripts
	oracleGenesis := oracletypes.DefaultGenesisState()

	oracleGenesis.Params.DataProviderRewardPerByte = sdk.NewDecCoinsFromCoins(Coin1geo)

	oracleGenesis.DataSources = getGenesisDataSources(dir)
	oracleGenesis.OracleScripts = getGenesisOracleScripts(dir)

	genesis[oracletypes.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")

	// Initialize the sim blockchain. We are ready for testing!
	app.InitChain(abci.RequestInitChain{
		ChainId:       chainID,
		Validators:    []abci.ValidatorUpdate{},
		AppStateBytes: stateBytes,
	})
	return app
}

// CreateTestInput creates a new test environment for unit tests.
// params[0] - activate;
// params[1] - fund pools;
func CreateTestInput(params ...bool) (*bandapp.BandApp, sdk.Context, me.Keeper) {
	app := NewSimApp("ODINCHAIN", log.NewNopLogger())
	ctx := app.NewContext(false, tmproto.Header{Height: app.LastBlockHeight()})
	if len(params) > 0 && params[0] {
		app.OracleKeeper.Activate(ctx, Validators[0].ValAddress)
		app.OracleKeeper.Activate(ctx, Validators[1].ValAddress)
		app.OracleKeeper.Activate(ctx, Validators[2].ValAddress)
	}

	if len(params) > 1 && params[1] {
		app.DistrKeeper.FundCommunityPool(ctx, DefaultCommunityPool, FeePoolProvider.Address)
		app.OracleKeeper.FundOraclePool(ctx, DefaultDataProvidersPool, OraclePoolProvider.Address)

		ctx = app.NewContext(false, tmproto.Header{})
	}
	return app, ctx, app.OracleKeeper
}
