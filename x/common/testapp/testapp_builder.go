package testapp

import (
	"encoding/json"
	bandapp "github.com/GeoDB-Limited/odin-core/app"
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type TestAppBuilder interface {
	Build(chainID string, stateBytes []byte, params ...bool) *bandapp.BandApp
	Codec() codec.Marshaler
	AddGenesis() TestAppBuilder
	UpdateModules(modulesGenesis map[string]json.RawMessage) TestAppBuilder

	GetAuthBuilder() *AuthBuilder
	GetStakingBuilder() *StakingBuilder
	GetBankBuilder() *BankBuilder
	GetOracleBuilder() *OracleBuilder

	SetAuthBuilder(*AuthBuilder)
	SetStakingBuilder(*StakingBuilder)
	SetBankBuilder(*BankBuilder)
	SetOracleBuilder(*OracleBuilder)
}

type testAppBuilder struct {
	app     *bandapp.BandApp
	genesis bandapp.GenesisState

	*AuthBuilder
	*StakingBuilder
	*BankBuilder
	*OracleBuilder
}

func (b *testAppBuilder) SetAuthBuilder(builder *AuthBuilder) {
	b.AuthBuilder = builder
}

func (b *testAppBuilder) SetStakingBuilder(builder *StakingBuilder) {
	b.StakingBuilder = builder
}

func (b *testAppBuilder) SetBankBuilder(builder *BankBuilder) {
	b.BankBuilder = builder
}

func (b *testAppBuilder) SetOracleBuilder(builder *OracleBuilder) {
	b.OracleBuilder = builder
}

func (b *testAppBuilder) GetAuthBuilder() *AuthBuilder {
	return b.AuthBuilder
}

func (b *testAppBuilder) GetStakingBuilder() *StakingBuilder {
	return b.StakingBuilder
}

func (b *testAppBuilder) GetBankBuilder() *BankBuilder {
	return b.BankBuilder
}

func (b *testAppBuilder) GetOracleBuilder() *OracleBuilder {
	return b.OracleBuilder
}

func NewTestAppBuilder(dir string, logger log.Logger) TestAppBuilder {
	builder := testAppBuilder{}

	db := dbm.NewMemDB()
	encCdc := bandapp.MakeEncodingConfig()
	builder.app = bandapp.NewBandApp(logger, db, nil, true, map[int64]bool{}, dir, 0, encCdc, EmptyAppOptions{}, false, 0)
	return &builder
}

func (b *testAppBuilder) Codec() codec.Marshaler {
	return b.app.AppCodec()
}

func (b *testAppBuilder) Build(chainID string, stateBytes []byte, params ...bool) *bandapp.BandApp {
	stateBytesNew := stateBytes
	if stateBytes == nil {
		stateBytesNew, _ = json.MarshalIndent(b.genesis, "", " ")
	}
	// Initialize the sim blockchain. We are ready for testing!
	b.app.InitChain(abci.RequestInitChain{
		ChainId:       chainID,
		Validators:    []abci.ValidatorUpdate{},
		AppStateBytes: stateBytesNew,
	})
	return b.app
}

func (b *testAppBuilder) AddGenesis() TestAppBuilder {
	b.genesis = bandapp.NewDefaultGenesisState()
	return b
}

func (b *testAppBuilder) UpdateModules(modulesGenesis map[string]json.RawMessage) TestAppBuilder {
	for k, v := range modulesGenesis {
		if v != nil {
			b.genesis[k] = v
		}
	}
	return b
}
