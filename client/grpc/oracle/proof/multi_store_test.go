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
						"data": "EtsCCgjAAAAAAAAAABKyAQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSEZQgOCllrsRWAEaCwgBGAEgASoDAAICIisIARIEAgQCIBohIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9IikIARIlBAgGIOmUjMDdzusCzGRZ3zJmMSlVGDSDYY1+5fEHZC6VKFNiICIpCAESJQYQBiDSkl4ALYGYlgBgemdjhdHoRN1cYdWFG23EPWvge8VONiAamQEKAfASBm9yYWNsZRoLCAEYASABKgMAAgIiKQgBEiUCBAIgYBmyUcOTZvAWUCIP0D+9tkdWu05jYdT58XI1aiWJin0gIikIARIlBAgGIOmUjMDdzusCzGRZ3zJmMSlVGDSDYY1+5fEHZC6VKFNiICIpCAESJQYQBiDSkl4ALYGYlgBgemdjhdHoRN1cYdWFG23EPWvge8VONiA="
					},
					{
						"type": "ics23:simple",
						"key": "b3JhY2xl",
						"data": "CtcBCgZvcmFjbGUSIGRXDu1yNxqG2TmOQcg+453c/SQ8HkqhZTbY7VWFcEqMGgkIARgBIAEqAQAiJwgBEgEBGiAHYjkK7NUeTCaEbwFUIDFagqKgxlJ11KSzBbQWhqlDbCInCAESAQEaIMUOwsPEUfdxKdW+jCZzjuvgJvh7RPH15SfaFIcPBDjdIicIARIBARogwBSM6GxVTpE8rPV49RCZFkeHuwpzO0FBOd/t4QmFlkIiJQgBEiEBnXL6Oo3x2e3JrICDcBxGw5E6s83QBN4QC+LeLdN6+R8="
					}
				]
			},
			"height": "3",
			"codespace": ""
		}
	}
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"CtcBCgZvcmFjbGUSIGRXDu1yNxqG2TmOQcg+453c/SQ8HkqhZTbY7VWFcEqMGgkIARgBIAEqAQAiJwgBEgEBGiAHYjkK7NUeTCaEbwFUIDFagqKgxlJ11KSzBbQWhqlDbCInCAESAQEaIMUOwsPEUfdxKdW+jCZzjuvgJvh7RPH15SfaFIcPBDjdIicIARIBARogwBSM6GxVTpE8rPV49RCZFkeHuwpzO0FBOd/t4QmFlkIiJQgBEiEBnXL6Oo3x2e3JrICDcBxGw5E6s83QBN4QC+LeLdN6+R8=",
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
		m.AuthToMintStoresMerkleHash,
		innerHash(
			innerHash(
				innerHash(
					leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
					m.ParamsStoreMerkleHash,
				),
				m.SlashingToStakingStoresMerkleHash,
			),
			m.TransferToUpgradeStoresMerkleHash,
		),
	)

	require.Equal(t, expectAppHash, apphash)
}
