package testapp

import (
	"encoding/json"
	bandapp "github.com/GeoDB-Limited/odin-core/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func CreateAppCustomValidators(accountsCount int, powers ...int) (*bandapp.BandApp, sdk.Context, TestAppBuilder) {
	builder := CreateDefaultGenesisApp(accountsCount)

	// staking
	stakingBuilder := NewStakingBuilder(len(powers), builder.GetStakingBuilder().BondDenom, powers...)
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = DefaultBondDenom
	validators, delegations := stakingBuilder.Build()
	initialSupply := countSupply(stakingParams.BondDenom, validators)
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	builder.SetStakingBuilder(stakingBuilder)

	// bank
	bankBuilder := NewBankBuilder(accountsCount, fillBalances(builder.GetAuthBuilder().Accounts), initialSupply)
	balances, totalSupply := bankBuilder.Build()
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultParams(), balances, totalSupply, []banktypes.Metadata{})
	builder.SetBankBuilder(bankBuilder)

	builder.UpdateModules(map[string]json.RawMessage{
		stakingtypes.ModuleName: builder.Codec().MustMarshalJSON(stakingGenesis),
		banktypes.ModuleName:    builder.Codec().MustMarshalJSON(bankGenesis),
	})

	app := builder.Build(DefaultChainID, nil)
	return app, app.NewContext(false, tmproto.Header{}), builder
}

func CreateAppCustomBalances(balancesRate ...int) (*bandapp.BandApp, sdk.Context, TestAppBuilder) {
	builder := CreateDefaultGenesisApp(len(balancesRate))

	balancesToFill := make([]sdk.Coins, 0, len(balancesRate))
	for _, br := range balancesRate {
		balancesToFill = append(balancesToFill, sdk.NewCoins(sdk.NewCoin(builder.GetStakingBuilder().BondDenom, sdk.TokensFromConsensusPower(int64(br)))))
	}

	bankBuilder := NewBankBuilder(len(balancesRate), fillBalances(builder.GetAuthBuilder().Accounts, balancesToFill...), sdk.NewCoins())
	balances, totalSupply := bankBuilder.Build()
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultParams(), balances, totalSupply, []banktypes.Metadata{})
	builder.SetBankBuilder(bankBuilder)

	builder.UpdateModules(map[string]json.RawMessage{
		banktypes.ModuleName: builder.Codec().MustMarshalJSON(bankGenesis),
	})

	app := builder.Build(DefaultChainID, nil)
	return app, app.NewContext(false, tmproto.Header{}), builder
}
