package common

import (
	"encoding/hex"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

func Test_FromHex(t *testing.T) {
	privRaw, _ := hex.DecodeString("4a4888d6e8eb96cdf96d774c982ed95f201a5068f1cabf95b2b18a7f37e1af5b")
	var priv secp256k1.PrivKey = privRaw

	pubKey := priv.PubKey()
	t.Log(pubKey)
	t.Log(pubKey.Address())
}
