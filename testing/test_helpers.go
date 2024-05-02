package testing

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	owasm "github.com/bandprotocol/go-owasm/api"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	"github.com/stretchr/testify/require"

	bandapp "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/pkg/filecache"
	"github.com/bandprotocol/chain/v2/testing/testdata"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// Account is a data structure to store key of test account.
type Account struct {
	PrivKey    cryptotypes.PrivKey
	PubKey     cryptotypes.PubKey
	Address    sdk.AccAddress
	ValAddress sdk.ValAddress
}

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

var (
	EmptyCoins          = sdk.Coins(nil)
	Coins1uband         = sdk.NewCoins(sdk.NewInt64Coin("uband", 1))
	Coins10uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	Coins11uband        = sdk.NewCoins(sdk.NewInt64Coin("uband", 11))
	Coins1000000uband   = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	Coins99999999uband  = sdk.NewCoins(sdk.NewInt64Coin("uband", 99999999))
	Coins100000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 100000000))
	BadCoins            = []sdk.Coin{{Denom: "uband", Amount: sdk.NewInt(-1)}}
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
var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   -1,
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
func (app *TestingApp) GetStakingKeeper() *stakingkeeper.Keeper {
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
	bandapp.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
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

// CreateTestApp creates a new test environment for unit tests.
func CreateTestApp(t *testing.T, autoActivate bool) (*TestingApp, sdk.Context) {
	t.Helper()

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

	isCheckTx := false
	app := setupWithGenesis(t, Validators, acc, ChainID, balances...)
	ctx := app.NewContext(isCheckTx, tmproto.Header{Height: app.LastBlockHeight()})
	if autoActivate {
		err := app.OracleKeeper.Activate(ctx, Validators[0].ValAddress)
		require.NoError(t, err)
		err = app.OracleKeeper.Activate(ctx, Validators[1].ValAddress)
		require.NoError(t, err)
		err = app.OracleKeeper.Activate(ctx, Validators[2].ValAddress)
		require.NoError(t, err)
	}

	return &TestingApp{app}, ctx
}

// setupWithGenesis sets up a BandApp for testing with a specified genesis state.
func setupWithGenesis(
	t *testing.T,
	valSet []Account,
	genAccs []authtypes.GenesisAccount,
	chainID string,
	balances ...banktypes.Balance,
) *bandapp.BandApp {
	t.Helper()

	app, genesisState, dir := setup(t, chainID, true)
	genesisState = generateGenesisState(t, app, dir, genesisState, valSet, genAccs, balances...)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	// Initialize the sim blockchain. We are ready for testing!
	app.InitChain(abci.RequestInitChain{
		ChainId:         chainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	return app
}

// setup initializes a BandApp instance for simulation testing.
func setup(t *testing.T, chainID string, withGenesis bool) (*bandapp.BandApp, bandapp.GenesisState, string) {
	t.Helper()

	// Set HomeFlag to a temp folder for simulation run.
	dir, err := os.MkdirTemp("", "bandd")
	if err != nil {
		panic(err)
	}
	db, _ := dbm.NewGoLevelDB("db", dir)

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = dir
	appOptions[server.FlagInvCheckPeriod] = 0

	snapshotDir := filepath.Join(dir, "data", "snapshots")
	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	snapshotOptions := snapshottypes.NewSnapshotOptions(
		1000,
		2,
	)

	app := bandapp.NewBandApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		appOptions,
		100,
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetChainID(chainID),
	)
	if withGenesis {
		return app, bandapp.NewDefaultGenesisState(), dir
	}

	return app, bandapp.GenesisState{}, dir
}

// SetupWithEmptyStore setup a instance with empty DB.
func SetupWithEmptyStore(t *testing.T, chainID string) *bandapp.BandApp {
	t.Helper()

	app, _, _ := setup(t, chainID, false)
	return app
}

// SetupWithGenesisValSet initializes a new TestingApp with a validator set and genesis accounts
// that also act as delegator. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the BandChain from first genesis
// account. A Nop logger is set in TestingApp.
func SetupWithGenesisValSet(
	t *testing.T,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	chainID string,
	balances ...banktypes.Balance,
) *TestingApp {
	t.Helper()

	app, genesisState, dir := setup(t, chainID, true)

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
		delegations = append(
			delegations,
			stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()),
		)
	}

	// set validators and delegations
	ps := stakingtypes.DefaultParams()
	ps.BondDenom = "uband"
	stakingGenesis := stakingtypes.NewGenesisState(ps, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin("uband", bondAmt))...)
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin("uband", bondAmt.Mul(sdk.NewInt(2)))},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	// Add genesis data sources and oracle scripts
	oracleGenesis := types.DefaultGenesisState()
	oracleGenesis.DataSources = generateDataSources(dir)
	oracleGenesis.OracleScripts = generateOracleScripts(dir)
	genesisState[types.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			ChainId:         chainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		ChainID:            chainID,
		Height:             app.LastBlockHeight() + 1,
		AppHash:            app.LastCommitID().Hash,
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
	}, Hash: app.LastCommitID().Hash})

	return &TestingApp{app}
}

// createArbitraryAccount generates a random Account using a provided random number generator.
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

// generateGenesisState initializes the genesis state of a BandApp for testing.
func generateGenesisState(
	t *testing.T,
	app *bandapp.BandApp,
	dir string,
	genesisState bandapp.GenesisState,
	valSet []Account,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) bandapp.GenesisState {
	t.Helper()

	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet))
	signingInfos := make([]slashingtypes.SigningInfo, 0, len(valSet))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet))
	bamt := []sdkmath.Int{Coins100000000uband[0].Amount, Coins1000000uband[0].Amount, Coins99999999uband[0].Amount}
	for idx, val := range valSet {
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
		signingInfos = append(
			signingInfos,
			slashingtypes.SigningInfo{Address: consAddr.String(), ValidatorSigningInfo: validatorSigningInfo},
		)
		delegations = append(
			delegations,
			stakingtypes.NewDelegation(genAccs[4+idx].GetAddress(), val.Address.Bytes(), sdk.OneDec()),
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
		Coins:   sdk.Coins{sdk.NewCoin("uband", sdk.NewInt(200999999))},
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
	oracleGenesis := types.DefaultGenesisState()
	oracleGenesis.DataSources = generateDataSources(dir)
	oracleGenesis.OracleScripts = generateOracleScripts(dir)
	genesisState[types.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)

	return genesisState
}

// generateDataSources generates a set of data sources for the BandApp.
func generateDataSources(homePath string) []types.DataSource {
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

// generateOracleScripts generates a set of oracle scripts for the BandApp.
func generateOracleScripts(homePath string) []types.OracleScript {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	OracleScripts = []types.OracleScript{{}} // 0th index should be ignored
	wasms := [][]byte{
		testdata.Wasm1, testdata.Wasm2, testdata.Wasm3, testdata.Wasm4, testdata.Wasm56(10), testdata.Wasm56(10000000), testdata.Wasm78(10), testdata.Wasm78(2000), testdata.Wasm9,
	}
	for idx := 0; idx < len(wasms); idx++ {
		idxStr := fmt.Sprintf("%d", idx+1)
		hash := fc.AddFile(testdata.Compile(wasms[idx]))
		OracleScripts = append(OracleScripts, types.NewOracleScript(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, "schema"+idxStr, "url"+idxStr,
		))
	}
	return OracleScripts[1:]
}

// GenTx generates a signed mock transaction.
func GenTx(
	gen client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums, accSeqs []uint64,
	priv ...cryptotypes.PrivKey,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}

// SignAndDeliver signs and delivers a transaction. No simulation occurs as the
// ibc testing package causes checkState and deliverState to diverge in block time.
func SignAndDeliver(
	t *testing.T, txCfg client.TxConfig, app *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := GenTx(
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin("uband", 2500)},
		DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{Header: header, Hash: header.AppHash})
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return gInfo, res, err
}
