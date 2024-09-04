package proof

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	ics23 "github.com/confio/ics23/go"
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
                        "data": "EtsCCgjAAAAAAAAAABKyAQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSEZQgOCllrsRWAEaCwgBGAEgASoDAAICIisIARIEAgQCIBohIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9IikIARIlBAgEIAXcsCtAWTDFinCgKJK8wl81ZbWfgSdKDpbP2Wf47EO4ICIpCAESJQYQBCCKaIo830SxbCL1Qk0rvRoWDuKiCxVidDRpX7yqJwZ6YCAamQEKAfASBm9yYWNsZRoLCAEYASABKgMAAgIiKQgBEiUCBAIgYBmyUcOTZvAWUCIP0D+9tkdWu05jYdT58XI1aiWJin0gIikIARIlBAgEIAXcsCtAWTDFinCgKJK8wl81ZbWfgSdKDpbP2Wf47EO4ICIpCAESJQYQBCCKaIo830SxbCL1Qk0rvRoWDuKiCxVidDRpX7yqJwZ6YCA="
                    },
                    {
                        "type": "ics23:simple",
                        "key": "b3JhY2xl",
                        "data": "CvoBCgZvcmFjbGUSIKPQ574oserr40+EoLhlrRgTGdrpgYViYEujG7W9rVBXGgkIARgBIAEqAQAiJQgBEiEBszziiEyGhp8/QWWNpdXqtweFbE6u087DWf5LXq0L+TIiJQgBEiEB3e3II3T8dhCPmDUxvJ1pJX5rydODMevJMyFcgoCkKOoiJQgBEiEBFpkM5/i+rVeuJ7iT04lRhhji55+3UobQFAvRoy5bAaIiJQgBEiEBTbPwtCi+YzaPpM9Zk+1Z73qE27/GXJESOD/JanvndjYiJwgBEgEBGiD3/25dJyNfa+hHLtAoLdyt2kECqzbUMTuiMuTW8f/kbA=="
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
		"CvoBCgZvcmFjbGUSIKPQ574oserr40+EoLhlrRgTGdrpgYViYEujG7W9rVBXGgkIARgBIAEqAQAiJQgBEiEBszziiEyGhp8/QWWNpdXqtweFbE6u087DWf5LXq0L+TIiJQgBEiEB3e3II3T8dhCPmDUxvJ1pJX5rydODMevJMyFcgoCkKOoiJQgBEiEBFpkM5/i+rVeuJ7iT04lRhhji55+3UobQFAvRoy5bAaIiJQgBEiEBTbPwtCi+YzaPpM9Zk+1Z73qE27/GXJESOD/JanvndjYiJwgBEgEBGiD3/25dJyNfa+hHLtAoLdyt2kECqzbUMTuiMuTW8f/kbA==",
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
