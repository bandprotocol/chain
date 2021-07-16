package testapp

import (
	"encoding/json"
	bandapp "github.com/GeoDB-Limited/odin-core/app"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"io/ioutil"
)

const (
	DefaultChainID       = "odin"
	DefaultAccountsCount = 10
)

func fillBalances(accounts []Account, values ...sdk.Coins) map[string]sdk.Coins {
	balances := make(map[string]sdk.Coins)
	for i, acc := range accounts {
		if len(values) < len(accounts) && i >= len(values) {
			balances[acc.Address.String()] = values[0]
		} else {
			balances[acc.Address.String()] = values[i]
		}
	}
	return balances
}

func countSupply(bondDenom string, validators stakingtypes.Validators) sdk.Coins {
	supply := sdk.NewCoins()
	for _, v := range validators {
		supply = supply.Add(sdk.NewCoin(bondDenom, v.Tokens))
	}
	return supply
}

func CreateDefaultGenesisApp(accountsCount int) TestAppBuilder {
	dir, err := ioutil.TempDir("", "bandd")
	if err != nil {
		panic(err)
	}
	viper.Set(cli.HomeFlag, dir)

	builder := NewTestAppBuilder(dir, log.NewNopLogger())

	// auth
	authBuilder := NewAuthBuilder(accountsCount)
	accounts := authBuilder.Build()
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), accounts)
	builder.SetAuthBuilder(authBuilder)

	// staking
	stakingBuilder := NewStakingBuilder(0, DefaultBondDenom)
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = stakingBuilder.BondDenom
	validators, delegations := stakingBuilder.Build()
	initialSupply := countSupply(stakingParams.BondDenom, validators)
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	builder.SetStakingBuilder(stakingBuilder)

	// bank
	bankBuilder := NewBankBuilder(accountsCount, fillBalances(authBuilder.Accounts, Coins10000000000loki), initialSupply)
	balances, totalSupply := bankBuilder.Build()
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultParams(), balances, totalSupply, []banktypes.Metadata{})
	builder.SetBankBuilder(bankBuilder)

	// oracle
	oracleBuilder := NewOracleBuilder(dir)
	oracleScripts, dataSources := oracleBuilder.Build()
	oracleGenesis := oracletypes.DefaultGenesisState()
	oracleGenesis.Params.DataProviderRewardPerByte = sdk.NewCoins(Coin1minigeo)

	oracleGenesis.OracleScripts = oracleScripts
	oracleGenesis.DataSources = dataSources
	builder.SetOracleBuilder(oracleBuilder)

	builder = builder.AddGenesis().UpdateModules(map[string]json.RawMessage{
		authtypes.ModuleName:    builder.Codec().MustMarshalJSON(authGenesis),
		stakingtypes.ModuleName: builder.Codec().MustMarshalJSON(stakingGenesis),
		banktypes.ModuleName:    builder.Codec().MustMarshalJSON(bankGenesis),
		oracletypes.ModuleName:  builder.Codec().MustMarshalJSON(oracleGenesis),
	})
	return builder
}

func CreateTestApp(params ...bool) (*bandapp.BandApp, sdk.Context) {
	// Set HomeFlag to a temp folder for simulation run.
	builder := CreateDefaultGenesisApp(DefaultAccountsCount)

	app := builder.Build(DefaultChainID, nil)
	ctx := app.NewContext(false, tmproto.Header{})

	if len(params) > 0 && params[0] {
		for _, v := range builder.GetStakingBuilder().Validators {
			err := app.OracleKeeper.Activate(ctx, v.GetOperator())
			if err != nil {
				panic(err)
			}
		}
	}

	if len(params) > 1 && params[1] {
		_ = app.DistrKeeper.FundCommunityPool(ctx, DefaultCommunityPool, FeePoolProvider.Address)
		_ = app.OracleKeeper.FundOraclePool(ctx, DefaultDataProvidersPool, OraclePoolProvider.Address)

		ctx = app.NewContext(false, tmproto.Header{})
	}

	return app, ctx
}
