package common

import (
	"bytes"
	"encoding/base64"
	"fmt"
	band "github.com/GeoDB-Limited/odin-core/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
)

func Test_FromBech32(t *testing.T) {
	config := sdk.GetConfig()
	accountPrefix := band.Bech32MainPrefix
	validatorPrefix := band.Bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator
	consensusPrefix := band.Bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus
	config.SetBech32PrefixForAccount(accountPrefix, accountPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(validatorPrefix, validatorPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(consensusPrefix, consensusPrefix+sdk.PrefixPublic)
	bech32ConsPub := "odinvalconspub1addwnpepqge86lvslkpfk0rlz0ah9tat0vntx8yele36hhfpflehfehydlutkvdvhfm"
	mustConsPub := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, bech32ConsPub)
	fmt.Println(mustConsPub.String())
	fmt.Println(mustConsPub.Type())
	bb := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, bb)
	encoder.Write(mustConsPub.Bytes())
	encoder.Close()
	fmt.Println(bb.String())
}
