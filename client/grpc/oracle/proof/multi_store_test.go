package proof

import (
	"testing"

	ics23 "github.com/confio/ics23/go"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

/*
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "response": {
      "code": 0,
      "log": "",
      "info": "",
      "index": "0",
      "key": "/wAAAAAAAAAB",
      "value": "Cglmcm9tX3NjYW4QJRoTAAAAAQAAAANCVEMAAAAAAAAAASABKAEwATgBQKanv5YGSK+nv5YGUAFaDAAAAAEAAAAAAABN/Q==",
      "proofOps": {
        "ops": [
          {
            "type": "ics23:iavl",
            "key": "/wAAAAAAAAAB",
            "data": "CqUGCgn/AAAAAAAAAAESRgoJZnJvbV9zY2FuECUaEwAAAAEAAAADQlRDAAAAAAAAAAEgASgBMAE4AUCmp7+WBkivp7+WBlABWgwAAAABAAAAAAAATf0aDAgBGAEgASoEAAKmByIsCAESBQIE7hwgGiEg1kJHijYj4LXp7yJdxjqkPo1FYMt7NsL1DgnLfcqHmrUiKwgBEicGCojlJiBruUg1auzUMVLTXFfhlxNtRoCleLipdQyUAQuvZhx+fyAiKwgBEicIGrCGOiDc8vTqp23PLI+oqAnnaDPJ5GfYbLHiNdUJp59OYzjl8CAiLQgBEgYKKrCGOiAaISDCrXQBFImjzEn6+MVNaUaQrZEY2AyxmmzqCavgq60qgSItCAESBgxKsIY6IBohIJtuvkFtWGPei1qDTLXIZD7Ua6ZZ6KQuABApK3urzrARIi4IARIHDooBsIY6IBohIDu44+RuyYrbayuQe9zSKS2kfo3iiazsZxs29UC6p4uiIiwIARIoEqoCpv9iICUQIPm6ilAKFaP6zfxWKH5au2ycdLVu3KdJnCVevEpqICIsCAESKBT6BKb/YiC5WKeSxCylF6XFCXxPMdhjeAe8hrcu0XlArfX8zxnWyCAiLggBEgcW+hSm/2IgGiEgpog44xpWyLZXFui26rY9TlSODHRdPE2IN6XCIDZb1y0iLAgBEigYuCOghmMgMUJfh9gYQaFEMjNlMtn9poGw2jOWSWg3a4p1463g/ZEgIi4IARIHGrhjoIZjIBohICMdVLMGqkr1rdg50jeCKxEttaaHJxXglzuX4wAu12YLIi8IARIIHOLcAaCGYyAaISC3g4DiugORMOVwBeE4oFver75z8GfW4MXztGjZJQ+I3SIvCAESCB7W0gOghmMgGiEgAQQIRNgVUpqsc1RS1ScymHSQNSk1tbLiT7T+iPX6at8iLwgBEgggstEHoIZjIBohIF1wseSwNKeXO1EDG4rW5r1qPAZbpcnYe4uSBWFK9INnIi8IARIIIrjcDqCGYyAaISAJFRWlM+9T5V8l/DexmzuAlvZ2iqB8//umaCVK27GMUQ=="
          },
          {
            "type": "ics23:simple",
            "key": "b3JhY2xl",
            "data": "Cv4BCgZvcmFjbGUSIMT+mWz9fL7gX5H1AiW7yVAXmGUkktBlOeF2V6DQGHBAGgkIARgBIAEqAQAiJwgBEgEBGiBbVrrKUHwejOlFfIHXQOt1ZoVSkjhvUdN4/Le2pMMIBCInCAESAQEaIP5aDHMJD2XUGAb1AQTE6wOzhWF4vcLj2B7mTftzbPaSIiUIARIhAdatdSzIzkclfVSgUbUYtE2UuzHMmNx1Zkh8sG7XoWWzIiUIARIhAVJHpAk7sdhIumCm11lfSzUrQRkm0DU2fR7tsB05bvlMIicIARIBARog5djvzlhwOubP4S4dKLAi3GdB5eja1e2Yt3kP9mOtxZg="
          }
        ]
      },
      "height": "811408",
      "codespace": ""
    }
  }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"Cv4BCgZvcmFjbGUSIMT+mWz9fL7gX5H1AiW7yVAXmGUkktBlOeF2V6DQGHBAGgkIARgBIAEqAQAiJwgBEgEBGiBbVrrKUHwejOlFfIHXQOt1ZoVSkjhvUdN4/Le2pMMIBCInCAESAQEaIP5aDHMJD2XUGAb1AQTE6wOzhWF4vcLj2B7mTftzbPaSIiUIARIhAdatdSzIzkclfVSgUbUYtE2UuzHMmNx1Zkh8sG7XoWWzIiUIARIhAVJHpAk7sdhIumCm11lfSzUrQRkm0DU2fR7tsB05bvlMIicIARIBARog5djvzlhwOubP4S4dKLAi3GdB5eja1e2Yt3kP9mOtxZg=",
	)

	var multistoreOps storetypes.CommitmentOp
	proof := &ics23.CommitmentProof{}
	err := proof.Unmarshal(data)
	require.Nil(t, err)

	multistoreOps = storetypes.NewSimpleMerkleCommitmentOp(key, proof)
	multistoreEp := multistoreOps.Proof.GetExist()
	require.NotNil(t, multistoreEp)

	var expectAppHash []byte
	expectAppHash, err = multistoreEp.Calculate()

	require.Nil(t, err)

	m := GetMultiStoreProof(multistoreEp)

	prefix := []byte{}
	prefix = append(prefix, 6)      // key length
	prefix = append(prefix, key...) // key to result of request #1
	prefix = append(prefix, 32)     // size of result hash must be 32

	apphash := innerHash(
		innerHash(
			m.AuthToFeegrantStoresMerkleHash,
			innerHash(
				m.GovToMintStoresMerkleHash,
				innerHash(
					innerHash(
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
						m.ParamsStoreMerkleHash,
					),
					m.SlashingToStakingStoresMerkleHash,
				),
			),
		),
		m.TransferToUpgradeStoresMerkleHash,
	)

	require.Equal(t, expectAppHash, apphash)
}
