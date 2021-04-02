package testapp

import (
	"encoding/json"
	"fmt"
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

	bandapp "github.com/bandprotocol/chain/app"
	"github.com/bandprotocol/chain/pkg/filecache"
	me "github.com/bandprotocol/chain/x/oracle/keeper"
	"github.com/bandprotocol/chain/x/oracle/types"
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
	Owner         Account
	Treasury      Account
	Alice         Account
	Bob           Account
	Carol         Account
	Validators    []Account
	DataSources   []types.DataSource
	OracleScripts []types.OracleScript
	OwasmVM       *owasm.Vm
)

// nolint
var (
	EmptyCoins          = sdk.Coins(nil)
	Coins1uband         = sdk.NewCoins(sdk.NewInt64Coin("uband", 1))
	Coins10uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	Coins11uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 11))
	Coins1000000uband   = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	Coins99999999uband  = sdk.NewCoins(sdk.NewInt64Coin("uband", 99999999))
	Coins100000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 100000000))
	Port1               = "port-1"
	Port2               = "port-2"
	Channel1            = "channel-1"
	Channel2            = "channel-2"
)

const (
	TestDefaultPrepareGas uint64 = 40000
	TestDefaultExecuteGas uint64 = 300000
)

func init() {
	bandapp.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
	r := rand.New(rand.NewSource(time.Now().Unix()))
	Owner = createArbitraryAccount(r)
	Treasury = createArbitraryAccount(r)
	Alice = createArbitraryAccount(r)
	Bob = createArbitraryAccount(r)
	Carol = createArbitraryAccount(r)
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

func getGenesisDataSources(homePath string) []types.DataSource {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	DataSources = []types.DataSource{{}} // 0th index should be ignored
	for idx := 0; idx < 5; idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile([]byte("code" + idxStr))
		DataSources = append(DataSources, types.NewDataSource(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, EmptyCoins, Owner.Address,
		))
	}
	return DataSources[1:]
}

func getGenesisOracleScripts(homePath string) []types.OracleScript {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	OracleScripts = []types.OracleScript{{}} // 0th index should be ignored
	wasms := [][]byte{
		Wasm1, Wasm2, Wasm3, Wasm4, Wasm56(10), Wasm56(10000000), Wasm78(10), Wasm78(2000), Wasm9,
	}
	for idx := 0; idx < len(wasms); idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile(compile(wasms[idx]))
		OracleScripts = append(OracleScripts, types.NewOracleScript(
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
	db := dbm.NewMemDB()
	encCdc := bandapp.MakeEncodingConfig()
	app := bandapp.NewBandApp(logger, db, nil, true, map[int64]bool{}, dir, 0, encCdc, EmptyAppOptions{}, false, 0)
	genesis := bandapp.NewDefaultGenesisState()
	acc := []authtypes.GenesisAccount{
		&authtypes.BaseAccount{Address: Owner.Address.String()},
		&authtypes.BaseAccount{Address: Alice.Address.String()},
		&authtypes.BaseAccount{Address: Bob.Address.String()},
		&authtypes.BaseAccount{Address: Carol.Address.String()},
		&authtypes.BaseAccount{Address: Validators[0].Address.String()},
		&authtypes.BaseAccount{Address: Validators[1].Address.String()},
		&authtypes.BaseAccount{Address: Validators[2].Address.String()},
	}
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), acc)
	genesis[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(Validators))
	bamt := []sdk.Int{Coins100000000uband[0].Amount, Coins1000000uband[0].Amount, Coins99999999uband[0].Amount}
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
			Status:            stakingtypes.Bonded,
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
	stakingParams.BondDenom = "uband"
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesis[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	// Fund seed accounts and validators with 1000000uband and 100000000uband initially.
	balances := []banktypes.Balance{
		{
			Address: Owner.Address.String(),
			Coins:   Coins1000000uband,
		},
		{Address: Alice.Address.String(), Coins: Coins1000000uband},
		{Address: Bob.Address.String(), Coins: Coins1000000uband},
		{Address: Carol.Address.String(), Coins: Coins1000000uband},
		{Address: Validators[0].Address.String(), Coins: Coins100000000uband},
		{Address: Validators[1].Address.String(), Coins: Coins100000000uband},
		{Address: Validators[2].Address.String(), Coins: Coins100000000uband},
	}
	totalSupply := sdk.NewCoins()
	for idx := 0; idx < len(balances)-len(validators); idx++ {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(balances[idx].Coins...)
	}
	for idx := 0; idx < len(validators); idx++ {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(balances[idx+len(balances)-len(validators)].Coins.Add(sdk.NewCoin("uband", bamt[idx]))...)
	}

	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	// Add genesis data sources and oracle scripts
	oracleGenesis := types.DefaultGenesisState()
	oracleGenesis.DataSources = getGenesisDataSources(dir)
	oracleGenesis.OracleScripts = getGenesisOracleScripts(dir)
	genesis[types.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)
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
func CreateTestInput(autoActivate bool) (*bandapp.BandApp, sdk.Context, me.Keeper) {
	app := NewSimApp("BANDCHAIN", log.NewNopLogger())
	ctx := app.NewContext(false, tmproto.Header{Height: app.LastBlockHeight()})
	if autoActivate {
		app.OracleKeeper.Activate(ctx, Validators[0].ValAddress)
		app.OracleKeeper.Activate(ctx, Validators[1].ValAddress)
		app.OracleKeeper.Activate(ctx, Validators[2].ValAddress)
	}
	return app, ctx, app.OracleKeeper
}
