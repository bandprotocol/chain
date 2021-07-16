package testapp

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"time"
)

type AuthBuilder struct {
	Accounts   []Account
	AccAddrMap map[string]Account
}

type StakingBuilder struct {
	ValAccounts []Account
	DelAccounts []Account
	Validators  stakingtypes.Validators
	Delegations stakingtypes.Delegations
	BondDenom   string
	Powers      []int
}

type BankBuilder struct {
	Balances    []banktypes.Balance
	TotalSupply sdk.Coins
	BalancesMap map[string]sdk.Coins
}

type OracleBuilder struct {
	homePath      string
	DataSources   []oracletypes.DataSource
	OracleScripts []oracletypes.OracleScript
}

func NewAuthBuilder(accCount int) *AuthBuilder {
	return &AuthBuilder{
		Accounts:   make([]Account, accCount),
		AccAddrMap: make(map[string]Account),
	}
}

func (b AuthBuilder) Build() []authtypes.GenesisAccount {
	res := make([]authtypes.GenesisAccount, len(b.Accounts))
	for i := range b.Accounts {
		b.Accounts[i] = createArbitraryAccount(RAND)
		b.AccAddrMap[b.Accounts[i].Address.String()] = b.Accounts[i]
		res[i] = &authtypes.BaseAccount{Address: b.Accounts[i].Address.String()}
	}
	return res
}

func NewStakingBuilder(valCount int, bondDenom string, powers ...int) *StakingBuilder {
	return &StakingBuilder{
		ValAccounts: make([]Account, valCount),
		Validators:  make(stakingtypes.Validators, valCount),
		BondDenom:   bondDenom,
		Powers:      powers,
	}
}

func (b StakingBuilder) AddDelegations(delegations ...sdk.Dec) {
	if len(b.Validators) != len(delegations) {
		panic("Invalid delegations count (should be equal to the number of validators)")
	}
	b.Delegations = make(stakingtypes.Delegations, 0, len(delegations))
	b.DelAccounts = make([]Account, 0, len(delegations))
	for i := range b.Delegations {
		b.DelAccounts[i] = createArbitraryAccount(RAND)
		b.Delegations[i] = stakingtypes.NewDelegation(b.ValAccounts[i].Address, b.DelAccounts[i].Address.Bytes(), sdk.OneDec())
	}
}

func (b StakingBuilder) Build() (stakingtypes.Validators, stakingtypes.Delegations) {
	for i := range b.Validators {
		b.ValAccounts[i] = createArbitraryAccount(RAND)
		pk, err := cryptocodec.FromTmPubKeyInterface(b.ValAccounts[i].PubKey)
		if err != nil {
			panic(err)
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			panic(err)
		}

		delegatorShares := sdk.ZeroDec()
		if b.Delegations != nil {
			delegatorShares = b.Delegations[i].GetShares()
		}
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(b.ValAccounts[i].Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Unbonded,
			Tokens:            sdk.TokensFromConsensusPower(int64(b.Powers[i])),
			DelegatorShares:   delegatorShares,
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		b.Validators[i] = validator
	}
	return b.Validators, b.Delegations
}

func NewBankBuilder(balCount int, balances map[string]sdk.Coins, initialSupply sdk.Coins) *BankBuilder {
	return &BankBuilder{
		Balances:    make([]banktypes.Balance, 0, balCount),
		BalancesMap: balances,
		TotalSupply: initialSupply,
	}
}

func (b BankBuilder) Build() ([]banktypes.Balance, sdk.Coins) {
	for addr, coins := range b.BalancesMap {
		b.Balances = append(b.Balances, banktypes.Balance{
			Address: addr,
			Coins:   coins,
		})
		b.TotalSupply = b.TotalSupply.Add(coins...)
	}
	return b.Balances, b.TotalSupply
}

func NewOracleBuilder(homePath string) *OracleBuilder {
	return &OracleBuilder{
		homePath: homePath,
	}
}

func (b OracleBuilder) Build() ([]oracletypes.OracleScript, []oracletypes.DataSource) {
	b.OracleScripts = getGenesisOracleScripts(b.homePath)
	b.DataSources = getGenesisDataSources(b.homePath)
	return b.OracleScripts, b.DataSources
}
