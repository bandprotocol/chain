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
                        "data": "EtsCCgjAAAAAAAAAABKyAQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSChQgOCllrsRWAEaCwgBGAEgASoDAAICIisIARIEAgQCIBohIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9IikIARIlBAYCIDj1W5djJSu7VMElKi6joXfpQCF9Wk/aOzZIKpCY6Ra2ICIpCAESJQYOBCCKaIo830SxbCL1Qk0rvRoWDuKiCxVidDRpX7yqJwZ6YCAamQEKAfASBm9yYWNsZRoLCAEYASABKgMAAgIiKQgBEiUCBAIgdi5vAneYh9+6Qc3ngEMlnwHOKVsdj3MA1Y18Q0Sfv1ggIikIARIlBAYCIDj1W5djJSu7VMElKi6joXfpQCF9Wk/aOzZIKpCY6Ra2ICIpCAESJQYOBCCKaIo830SxbCL1Qk0rvRoWDuKiCxVidDRpX7yqJwZ6YCA="
                    },
                    {
                        "type": "ics23:simple",
                        "key": "b3JhY2xl",
                        "data": "CvoBCgZvcmFjbGUSIKOu1UEe8Quc2EL/1+wv12JvlreYOeYIttnz9IEULBF5GgkIARgBIAEqAQAiJQgBEiEBszziiEyGhp8/QWWNpdXqtweFbE6u087DWf5LXq0L+TIiJQgBEiEB3e3II3T8dhCPmDUxvJ1pJX5rydODMevJMyFcgoCkKOoiJQgBEiEB8sFioDJU8QbQ8eoWJa4ohiwrc28YRUAmyvF9anblxgAiJQgBEiEBXRVe5Pdw6aqG0QqzJ4R7u9yNwbcffW/0DrPxR6fSMMQiJwgBEgEBGiCRUnPfxybN1OikJ39IxNDtLuhkE5ECBhk0Y7pLsCJcLQ=="
                    }
                ]
            },
            "height": "2813",
            "codespace": ""
        }
    }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"CvoBCgZvcmFjbGUSIKOu1UEe8Quc2EL/1+wv12JvlreYOeYIttnz9IEULBF5GgkIARgBIAEqAQAiJQgBEiEBszziiEyGhp8/QWWNpdXqtweFbE6u087DWf5LXq0L+TIiJQgBEiEB3e3II3T8dhCPmDUxvJ1pJX5rydODMevJMyFcgoCkKOoiJQgBEiEB8sFioDJU8QbQ8eoWJa4ohiwrc28YRUAmyvF9anblxgAiJQgBEiEBXRVe5Pdw6aqG0QqzJ4R7u9yNwbcffW/0DrPxR6fSMMQiJwgBEgEBGiCRUnPfxybN1OikJ39IxNDtLuhkE5ECBhk0Y7pLsCJcLQ==",
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
			m.AuthToEvidenceStoresMerkleHash,
			innerHash(
				m.FeegrantToGroupStoresMerkleHash,
				innerHash(
					m.IbcToIcahostStoresMerkleHash,
					innerHash(
						m.MintStoreMerkleHash,
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
					),
				),
			),
		),
		m.ParamsToUpgradeStoresMerkleHash,
	)

	require.Equal(t, expectAppHash, apphash)
}
