package proof

import (
	"testing"

	ics23 "github.com/confio/ics23/go"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
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
            "data": "EukCCgjAAAAAAAAAABLAAQoMAFJvbGxpbmdTZWVkEiAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD13RoLCAEYASABKgMAAgQiKwgBEgQCBAQgGiEg63Obsi9It/MFOpC6K6T+B/qyYsrfhmRIlWXFD/UFuL0iKQgBEiUEBgQgOPVbl2MlK7tUwSUqLqOhd+lAIX1aT9o7NkgqkJjpFrYgIikIARIlBg4EIIpoijzfRLFsIvVCTSu9GhYO4qILFWJ0NGlfvKonBnpgIBqZAQoB8BIGb3JhY2xlGgsIARgBIAEqAwACAiIpCAESJQIEBCBYKdm973ulzdqUE+jnuc4DZHxw3sjFNe5vcSIW8V+YriAiKQgBEiUEBgQgOPVbl2MlK7tUwSUqLqOhd+lAIX1aT9o7NkgqkJjpFrYgIikIARIlBg4EIIpoijzfRLFsIvVCTSu9GhYO4qILFWJ0NGlfvKonBnpgIA=="
          },
          {
            "type": "ics23:simple",
            "key": "b3JhY2xl",
            "data": "CvwBCgZvcmFjbGUSIAvDBkyKB8SaeNb7cGjhsWvPuddSyVYD9YcDwBR1XwurGgkIARgBIAEqAQAiJQgBEiEB9VTw+0c6LUfGylARB2nWup+a0Lz0aC9eeJV6X5WGNNgiJwgBEgEBGiBF9TB9QwOUKdGmWOos/jVWy3EgicQNb+oMRkFqtXgauCIlCAESIQFrz6rMKwx5aOJG+y8vmBEz/VRqaHlgjqkS4ZLcVkrS8yIlCAESIQGpKs2CX3ZSYSnuK+8DM0QNNX6m1D6h39qoUzEUlgljrSInCAESAQEaICBo1MHdfzcpuU+TfQKKlkDcXlZOWyByO22ZEEhqErrz"
          }
        ]
      },
      "height": "2",
      "codespace": ""
    }
  }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"CvwBCgZvcmFjbGUSIAvDBkyKB8SaeNb7cGjhsWvPuddSyVYD9YcDwBR1XwurGgkIARgBIAEqAQAiJQgBEiEB9VTw+0c6LUfGylARB2nWup+a0Lz0aC9eeJV6X5WGNNgiJwgBEgEBGiBF9TB9QwOUKdGmWOos/jVWy3EgicQNb+oMRkFqtXgauCIlCAESIQFrz6rMKwx5aOJG+y8vmBEz/VRqaHlgjqkS4ZLcVkrS8yIlCAESIQGpKs2CX3ZSYSnuK+8DM0QNNX6m1D6h39qoUzEUlgljrSInCAESAQEaICBo1MHdfzcpuU+TfQKKlkDcXlZOWyByO22ZEEhqErrz",
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
				m.GlobalfeeToIbccoreStoresMerkleHash,
				innerHash(
					m.IcahostToMintStoresMerkleHash,
					innerHash(
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
						m.ParamsStoreMerkleHash,
					),
				),
			),
		),
		m.SlashingToUpgradeStoresMerkleHash,
	)

	require.Equal(t, expectAppHash, apphash)
}
