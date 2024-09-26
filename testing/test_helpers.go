package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	cosmosdb "github.com/cosmos/cosmos-db"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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

func init() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	Owner = createArbitraryAccount(r)
	Treasury = createArbitraryAccount(r)
	FeePayer = createArbitraryAccount(r)
	Alice = createArbitraryAccount(r)
	Bob = createArbitraryAccount(r)
	Carol = createArbitraryAccount(r)
	MissedValidator = createArbitraryAccount(r)
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

// createArbitraryAccount generates a random Account using a provided random number generator.
func createArbitraryAccount(r *rand.Rand) Account {
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
	oracleGenesis.DataSources = generateDataSources(dir)
	oracleGenesis.OracleScripts = generateOracleScripts(dir)
	genesisState[oracletypes.ModuleName] = app.AppCodec().MustMarshalJSON(oracleGenesis)

	return genesisState
}

// generateDataSources generates a set of data sources for the BandApp.
func generateDataSources(homePath string) []oracletypes.DataSource {
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

// generateOracleScripts generates a set of oracle scripts for the BandApp.
func generateOracleScripts(homePath string) []oracletypes.OracleScript {
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

func SetupWithCustomHomeAndChainId(isCheckTx bool, dir, chainId string) *band.BandApp {
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
		100,
		baseapp.SetChainID(chainId),
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
				ChainId:         chainId,
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
			100,
		)

		g := GenesisStateWithValSet(app, dir)
		return app, g
	}
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

	signMode, err := authsigning.APISignModeToInternal(gen.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

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

	txBuilder := gen.NewTxBuilder()

	if err = txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	if err = txBuilder.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}

		sig, err := tx.SignWithPrivKey(context.Background(), signMode, signerData, txBuilder, p, gen, accSeqs[i])
		if err != nil {
			panic(err)
		}

		sigs[i] = sig
	}

	if err = txBuilder.SetSignatures(sigs...); err != nil {
		panic(err)
	}

	return txBuilder.GetTx(), nil
}
