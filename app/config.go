package band

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBech32AddressPrefixesAndBip44CoinTypeAndSeal sets the global Bech32 prefixes and HD wallet coin type and seal config.
func SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(config *sdk.Config) {
	accountPrefix := Bech32MainPrefix
	validatorPrefix := Bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator
	consensusPrefix := Bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus
	config.SetCoinType(Bip44CoinType)
	config.SetBech32PrefixForAccount(accountPrefix, accountPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(validatorPrefix, validatorPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(consensusPrefix, consensusPrefix+sdk.PrefixPublic)

	config.Seal()
}
