package testing

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	cosmosdb "github.com/cosmos/cosmos-db"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	owasm "github.com/bandprotocol/go-owasm/api"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/filecache"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

// Account is a data structure to store key of test account.
type Account struct {
	PrivKey    cryptotypes.PrivKey
	PubKey     cryptotypes.PubKey
	Address    sdk.AccAddress
	ValAddress sdk.ValAddress
}

var (
	Owner           Account
	Treasury        Account
	FeePayer        Account
	Alice           Account
	Bob             Account
	Carol           Account
	MissedValidator Account
	Validators      []Account
	DataSources     []oracletypes.DataSource
	OracleScripts   []oracletypes.OracleScript
	OwasmVM         *owasm.Vm
)

var (
	EmptyCoins          = sdk.Coins(nil)
	Coins1uband         = sdk.NewCoins(sdk.NewInt64Coin("uband", 1))
	Coins10uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	Coins11uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 11))
	Coins1000000uband   = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	Coins99999999uband  = sdk.NewCoins(sdk.NewInt64Coin("uband", 99999999))
	Coins100000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 100000000))
	BadCoins            = []sdk.Coin{{Denom: "uband", Amount: math.NewInt(-1)}}
)

const (
	ChainID               string = "BANDCHAIN"
	Port1                        = "port-1"
	Port2                        = "port-2"
	Channel1                     = "channel-1"
	Channel2                     = "channel-2"
	TestDefaultPrepareGas uint64 = 40000
	TestDefaultExecuteGas uint64 = 300000
	DefaultGenTxGas              = 1000000
)

// DefaultConsensusParams defines the default Tendermint consensus params used in TestingApp.
var DefaultConsensusParams = &cmtproto.ConsensusParams{
	Block: &cmtproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   -1,
	},
	Evidence: &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		// MaxBytes:        10000,
	},
	Validator: &cmtproto.ValidatorParams{
		PubKeyTypes: []string{
			cmttypes.ABCIPubKeyTypeSecp256k1,
		},
	},
}

func init() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	Owner = CreateArbitraryAccount(r)
	Treasury = CreateArbitraryAccount(r)
	FeePayer = CreateArbitraryAccount(r)
	Alice = CreateArbitraryAccount(r)
	Bob = CreateArbitraryAccount(r)
	Carol = CreateArbitraryAccount(r)
	MissedValidator = CreateArbitraryAccount(r)
	for i := 0; i < 3; i++ {
		Validators = append(Validators, CreateArbitraryAccount(r))
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

// CreateArbitraryAccount generates a random Account using a provided random number generator.
func CreateArbitraryAccount(r *rand.Rand) Account {
	privkeySeed := make([]byte, 12)
	_, _ = r.Read(privkeySeed)
	privKey := secp256k1.GenPrivKeyFromSecret(privkeySeed)
	return Account{
		PrivKey:    privKey,
		PubKey:     privKey.PubKey(),
		Address:    sdk.AccAddress(privKey.PubKey().Address()),
		ValAddress: sdk.ValAddress(privKey.PubKey().Address()),
	}
}

func GenesisStateWithValSet(app *band.BandApp, dir string) band.GenesisState {
	genAccs := []authtypes.GenesisAccount{
		&authtypes.BaseAccount{Address: Owner.Address.String()},
		&authtypes.BaseAccount{Address: FeePayer.Address.String()},
		&authtypes.BaseAccount{Address: Alice.Address.String()},
		&authtypes.BaseAccount{Address: Bob.Address.String()},
		&authtypes.BaseAccount{Address: Carol.Address.String()},
		&authtypes.BaseAccount{Address: MissedValidator.Address.String()},
		&authtypes.BaseAccount{Address: Validators[0].Address.String()},
		&authtypes.BaseAccount{Address: Validators[1].Address.String()},
		&authtypes.BaseAccount{Address: Validators[2].Address.String()},
	}

	balances := []banktypes.Balance{
		{
			Address: Owner.Address.String(),
			Coins:   Coins1000000uband,
		},
		{Address: FeePayer.Address.String(), Coins: Coins100000000uband},
		{Address: Alice.Address.String(), Coins: Coins1000000uband},
		{Address: Bob.Address.String(), Coins: Coins1000000uband},
		{Address: Carol.Address.String(), Coins: Coins1000000uband},
		{Address: MissedValidator.Address.String(), Coins: Coins100000000uband},
		{Address: Validators[0].Address.String(), Coins: Coins100000000uband},
		{Address: Validators[1].Address.String(), Coins: Coins100000000uband},
		{Address: Validators[2].Address.String(), Coins: Coins100000000uband},
	}

	genesisState := band.NewDefaultGenesisState(app.AppCodec())
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(Validators))
	signingInfos := make([]slashingtypes.SigningInfo, 0, len(Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(Validators))
	bamt := []math.Int{Coins100000000uband[0].Amount, Coins1000000uband[0].Amount, Coins99999999uband[0].Amount}
	for idx, val := range Validators {
		pkAny, _ := codectypes.NewAnyWithValue(val.PubKey)
		validator := stakingtypes.Validator{
			OperatorAddress: val.ValAddress.String(),
			ConsensusPubkey: pkAny,
			Jailed:          false,
			Status:          stakingtypes.Bonded,
			Tokens:          bamt[idx],
			DelegatorShares: math.LegacyOneDec(),
			Description:     stakingtypes.Description{},
			UnbondingHeight: int64(0),
			UnbondingTime:   time.Unix(0, 0).UTC(),
			Commission: stakingtypes.NewCommission(
				math.LegacyZeroDec(),
				math.LegacyZeroDec(),
				math.LegacyZeroDec(),
			),
			MinSelfDelegation: math.ZeroInt(),
		}
		consAddr, err := validator.GetConsAddr()
		validatorSigningInfo := slashingtypes.NewValidatorSigningInfo(consAddr, 0, 0, time.Unix(0, 0), false, 0)
		if err != nil {
			panic(err)
		}
		validators = append(validators, validator)
		signingInfos = append(
			signingInfos,
			slashingtypes.SigningInfo{
				Address:              sdk.ConsAddress(consAddr).String(),
				ValidatorSigningInfo: validatorSigningInfo,
			},
		)
		delegations = append(
			delegations,
			stakingtypes.NewDelegation(
				genAccs[4+idx].GetAddress().String(),
				val.ValAddress.String(),
				math.LegacyOneDec(),
			),
		)
	}
	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = "uband"
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	slashingParams := slashingtypes.DefaultParams()
	slashingGenesis := slashingtypes.NewGenesisState(slashingParams, signingInfos, nil)
	genesisState[slashingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(slashingGenesis)

	totalSupply := sdk.NewCoins()
	for idx := 0; idx < len(balances)-len(validators); idx++ {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(balances[idx].Coins...)
	}
	for idx := 0; idx < len(validators); idx++ {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(
			balances[idx+len(balances)-len(validators)].Coins.Add(sdk.NewCoin("uband", bamt[idx]))...)
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin("uband", math.NewInt(200999999))},
	})

	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	// Add genesis data sources and oracle scripts
	oracleGenesis := oracletypes.DefaultGenesisState()
	oracleGenesis.DataSources = GenerateDataSources(dir)
	oracleGenesis.OracleScripts = GenerateOracleScripts(dir)
	genesisState[oracletypes.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)

	return genesisState
}

// GenerateDataSources generates a set of data sources for the BandApp.
func GenerateDataSources(homePath string) []oracletypes.DataSource {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	DataSources = []oracletypes.DataSource{{}} // 0th index should be ignored
	for idx := 0; idx < 5; idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile([]byte("code" + idxStr))
		DataSources = append(DataSources, oracletypes.NewDataSource(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, Coins1000000uband, Treasury.Address,
		))
	}
	return DataSources[1:]
}

// GenerateOracleScripts generates a set of oracle scripts for the BandApp.
func GenerateOracleScripts(homePath string) []oracletypes.OracleScript {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	OracleScripts = []oracletypes.OracleScript{{}} // 0th index should be ignored
	wasms := [][]byte{
		testdata.Wasm1, testdata.Wasm2, testdata.Wasm3, testdata.Wasm4, testdata.Wasm56(10), testdata.Wasm56(10000000), testdata.Wasm78(10), testdata.Wasm78(2000), testdata.Wasm9,
	}
	for idx := 0; idx < len(wasms); idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile(testdata.Compile(wasms[idx]))
		OracleScripts = append(OracleScripts, oracletypes.NewOracleScript(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, "schema"+idxStr, "url"+idxStr,
		))
	}
	return OracleScripts[1:]
}

// SetupWithCustomHome initializes a new BandApp with a custom home directory
func SetupWithCustomHome(isCheckTx bool, dir string) *band.BandApp {
	return SetupWithCustomHomeAndChainId(isCheckTx, dir, ChainID)
}

func SetupWithCustomHomeAndChainId(isCheckTx bool, dir, chainID string) *band.BandApp {
	db := cosmosdb.NewMemDB()

	snapshotDir := filepath.Join(dir, "data", "snapshots")
	snapshotDB, err := cosmosdb.NewDB("metadata", cosmosdb.GoLevelDBBackend, snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	app := band.NewBandApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		dir,
		sims.EmptyAppOptions{},
		[]wasmkeeper.Option{},
		100,
		baseapp.SetChainID(chainID),
		baseapp.SetSnapshot(snapshotStore, snapshottypes.SnapshotOptions{KeepRecent: 2}),
	)
	if !isCheckTx {
		genesisState := GenesisStateWithValSet(app, dir)
		defaultGenesisStatebytes, err := json.Marshal(genesisState)
		if err != nil {
			panic(err)
		}

		_, err = app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   defaultGenesisStatebytes,
				ChainId:         chainID,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	return app
}

func CreateTestingAppFn(t testing.TB) func() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		dir := testutil.GetTempDir(t)
		app := band.NewBandApp(
			log.NewNopLogger(),
			cosmosdb.NewMemDB(),
			nil,
			true,
			map[int64]bool{},
			dir,
			sims.EmptyAppOptions{},
			[]wasmkeeper.Option{},
			100,
		)

		g := GenesisStateWithValSet(app, dir)
		return app, g
	}
}
