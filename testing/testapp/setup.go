package testapp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	ibckeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/bandprotocol/chain/v2/pkg/filecache"
	owasm "github.com/bandprotocol/go-owasm/api"

	bandapp "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// Account is a data structure to store key of test account.
type Account struct {
	PrivKey    cryptotypes.PrivKey
	PubKey     cryptotypes.PubKey
	Address    sdk.AccAddress
	ValAddress sdk.ValAddress
}

// nolint
var (
	Owner         Account
	Treasury      Account
	FeePayer      Account
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
	BadCoins            = []sdk.Coin{{Denom: "uband", Amount: sdk.NewInt(-1)}}
	Port1               = "port-1"
	Port2               = "port-2"
	Channel1            = "channel-1"
	Channel2            = "channel-2"
)

const (
	TestDefaultPrepareGas uint64 = 40000
	TestDefaultExecuteGas uint64 = 300000
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// SimApp testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		// MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeSecp256k1,
		},
	},
}

type TestingApp struct {
	*bandapp.BandApp
}

func (app *TestingApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetStakingKeeper implements the TestingApp interface.
func (app *TestingApp) GetStakingKeeper() stakingkeeper.Keeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *TestingApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *TestingApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *TestingApp) GetTxConfig() client.TxConfig {
	return bandapp.MakeEncodingConfig().TxConfig
}

func init() {
	bandapp.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
	r := rand.New(rand.NewSource(time.Now().Unix()))
	Owner = createArbitraryAccount(r)
	Treasury = createArbitraryAccount(r)
	FeePayer = createArbitraryAccount(r)
	Alice = createArbitraryAccount(r)
	Bob = createArbitraryAccount(r)
	Carol = createArbitraryAccount(r)
	for i := 0; i < 3; i++ {
		Validators = append(Validators, createArbitraryAccount(r))
	}

	// Sorted list of validators is needed for ibctest when signing a commit block
	sort.Slice(Validators, func(i, j int) bool {
		return Validators[i].PubKey.Address().String() < Validators[j].PubKey.Address().String()
	})

	owasmVM, err := owasm.NewVm(10)
	if err != nil {
		panic(err)
	}
	OwasmVM = owasmVM
}

func createArbitraryAccount(r *rand.Rand) Account {
	privkeySeed := make([]byte, 12)
	r.Read(privkeySeed)
	privKey := secp256k1.GenPrivKeyFromSecret(privkeySeed)
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
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, Coins1000000uband, Treasury.Address,
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

// NewTestApp creates instance of our app using in test.
func NewTestApp(chainID string, logger log.Logger) *TestingApp {
	// Set HomeFlag to a temp folder for simulation run.
	dir, err := ioutil.TempDir("", "bandd")
	if err != nil {
		panic(err)
	}
	db := dbm.NewMemDB()
	encCdc := bandapp.MakeEncodingConfig()
	app := &TestingApp{
		BandApp: bandapp.NewBandApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, dir, 0, encCdc, EmptyAppOptions{}, false, 0),
	}
	genesis := bandapp.NewDefaultGenesisState()
	acc := []authtypes.GenesisAccount{
		&authtypes.BaseAccount{Address: Owner.Address.String()},
		&authtypes.BaseAccount{Address: FeePayer.Address.String()},
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
	signingInfos := make([]slashingtypes.SigningInfo, 0, len(Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(Validators))
	bamt := []sdk.Int{Coins100000000uband[0].Amount, Coins1000000uband[0].Amount, Coins99999999uband[0].Amount}
	// bondAmt := sdk.NewInt(1000000)
	for idx, val := range Validators {
		tmpk, err := cryptocodec.ToTmPubKeyInterface(val.PubKey)
		if err != nil {
			panic(err)
		}
		pk, err := cryptocodec.FromTmPubKeyInterface(tmpk)
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
		consAddr, err := validator.GetConsAddr()
		validatorSigningInfo := slashingtypes.NewValidatorSigningInfo(consAddr, 0, 0, time.Unix(0, 0), false, 0)
		if err != nil {
			panic(err)
		}
		validators = append(validators, validator)
		signingInfos = append(signingInfos, slashingtypes.SigningInfo{Address: consAddr.String(), ValidatorSigningInfo: validatorSigningInfo})
		delegations = append(delegations, stakingtypes.NewDelegation(acc[4+idx].GetAddress(), val.Address.Bytes(), sdk.OneDec()))
	}
	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = "uband"
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesis[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	slashingParams := slashingtypes.DefaultParams()
	slashingGenesis := slashingtypes.NewGenesisState(slashingParams, signingInfos, nil)
	genesis[slashingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(slashingGenesis)

	// Fund seed accounts and validators with 1000000uband and 100000000uband initially.
	balances := []banktypes.Balance{
		{
			Address: Owner.Address.String(),
			Coins:   Coins1000000uband,
		},
		{Address: FeePayer.Address.String(), Coins: Coins100000000uband},
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
func CreateTestInput(autoActivate bool) (*TestingApp, sdk.Context, keeper.Keeper) {
	app := NewTestApp("BANDCHAIN", log.NewNopLogger())
	ctx := app.NewContext(false, tmproto.Header{Height: app.LastBlockHeight()})
	if autoActivate {
		app.OracleKeeper.Activate(ctx, Validators[0].ValAddress)
		app.OracleKeeper.Activate(ctx, Validators[1].ValAddress)
		app.OracleKeeper.Activate(ctx, Validators[2].ValAddress)
	}
	return app, ctx, app.OracleKeeper
}

func setup(withGenesis bool, invCheckPeriod uint) (*TestingApp, bandapp.GenesisState, string) {
	dir, err := ioutil.TempDir("", "bandibc")
	if err != nil {
		panic(err)
	}
	db := dbm.NewMemDB()
	encCdc := bandapp.MakeEncodingConfig()
	app := &TestingApp{
		BandApp: bandapp.NewBandApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, dir, 0, encCdc, EmptyAppOptions{}, false, 0),
	}
	if withGenesis {
		return app, bandapp.NewDefaultGenesisState(), dir
	}
	return app, bandapp.GenesisState{}, dir
}

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *TestingApp {
	app, genesisState, dir := setup(true, 5)
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.NewInt(1000000)

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))

	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))...)
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	// Add genesis data sources and oracle scripts
	oracleGenesis := types.DefaultGenesisState()
	oracleGenesis.DataSources = getGenesisDataSources(dir)
	oracleGenesis.OracleScripts = getGenesisOracleScripts(dir)
	genesisState[types.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:             app.LastBlockHeight() + 1,
		AppHash:            app.LastCommitID().Hash,
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
	}, Hash: app.LastCommitID().Hash})

	return app
}

// SimAppChainID hardcoded chainID for simulation
const (
	DefaultGenTxGas = 1000000
	SimAppChainID   = "simulation-app"
)

// SignAndDeliver signs and delivers a transaction. No simulation occurs as the
// ibc testing package causes checkState and deliverState to diverge in block time.
func SignAndDeliver(
	t *testing.T, txCfg client.TxConfig, app *bam.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {

	tx, err := helpers.GenTx(
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		helpers.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{Header: header, Hash: header.AppHash})
	gInfo, res, err := app.Deliver(txCfg.TxEncoder(), tx)

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return gInfo, res, err
}
