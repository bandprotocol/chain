package proof

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/tmhash"
	ics23 "github.com/confio/ics23/go"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
)

/*
query at localhost:26657/abci_query?path="/store/oracle/key"&data=0xc000000000000000&prove=true
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "response": {
      "code": 0,
      "log": "",
      "info": "",
      "index": "0",
      "key": "wAAAAAAAAAA=",
      "value": null,
      "proofOps": {
        "ops": [
          {
            "type": "ics23:iavl",
            "key": "wAAAAAAAAAA=",
            "data": "EtsCCgjAAAAAAAAAABKyAQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSChQgOCllrsRWAEaCwgBGAEgASoDAAICIisIARIEAgQCIBohIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9IikIARIlBAYCIDj1W5djJSu7VMElKi6joXfpQCF9Wk/aOzZIKpCY6Ra2ICIpCAESJQYOCCCtJh9hnFH+nCXUc/vv3W7QfQEwx0Ubf2Lxg+N9C127iCAamQEKAfASBm9yYWNsZRoLCAEYASABKgMAAgIiKQgBEiUCBAIgdi5vAneYh9+6Qc3ngEMlnwHOKVsdj3MA1Y18Q0Sfv1ggIikIARIlBAYCIDj1W5djJSu7VMElKi6joXfpQCF9Wk/aOzZIKpCY6Ra2ICIpCAESJQYOCCCtJh9hnFH+nCXUc/vv3W7QfQEwx0Ubf2Lxg+N9C127iCA="
          },
          {
            "type": "ics23:simple",
            "key": "b3JhY2xl",
            "data": "Cv4BCgZvcmFjbGUSIKXZZIz6tibuBT/XZFlzvARMvAYUk9E9+irQmktTAc7TGgkIARgBIAEqAQAiJQgBEiEBBTl35L9FDWYOYAUJrkHLr8meoEluSU1g2NfAFCU8a3YiJwgBEgEBGiBV7R+2oGZxTGgzi2R7nTnYoaOt3XoszZUxlOmUxW4YCyInCAESAQEaIOzLegRHxMW0Unv0rFaJyJMuGuReOkAl0zrnevlu2LQ+IicIARIBARog35ba8pwX7BvU9VCfJzDMfY/V9uM4X7KOGVmjRaQvfCEiJQgBEiEBmk6jPft10k6/2duBuAdRbAGsRYdAoafrTV4fKNydPoA="
          }
        ]
      },
      "height": "4",
      "codespace": ""
    }
  }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"Cv4BCgZvcmFjbGUSIKXZZIz6tibuBT/XZFlzvARMvAYUk9E9+irQmktTAc7TGgkIARgBIAEqAQAiJQgBEiEBBTl35L9FDWYOYAUJrkHLr8meoEluSU1g2NfAFCU8a3YiJwgBEgEBGiBV7R+2oGZxTGgzi2R7nTnYoaOt3XoszZUxlOmUxW4YCyInCAESAQEaIOzLegRHxMW0Unv0rFaJyJMuGuReOkAl0zrnevlu2LQ+IicIARIBARog35ba8pwX7BvU9VCfJzDMfY/V9uM4X7KOGVmjRaQvfCEiJQgBEiEBmk6jPft10k6/2duBuAdRbAGsRYdAoafrTV4fKNydPoA=",
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
		m.AuthToIcahostStoresMerkleHash,
		innerHash(
			innerHash(
				innerHash(
					innerHash(
						m.MintStoreMerkleHash,
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
					),
					m.ParamsToRollingseedStoresMerkleHash,
				),
				m.SlashingToTssStoresMerkleHash,
			),
			m.UpgradeStoreMerkleHash,
		),
	)

	require.Equal(t, expectAppHash, apphash)
}
