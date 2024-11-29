package proof

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/tmhash"

	ics23 "github.com/cosmos/ics23/go"

	storetypes "cosmossdk.io/store/types"
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
                        "data": "Eo8DCgjAAAAAAAAAABK1AQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSEZQgOCllrsRWAEaCwgBGAEgASoDAAICIiwIARIFBAaIAiAaISBtvVJ2ZZMSBA4nAMaNXfsRtGLoznC27KcRDwrP60ri1iIqCAESJgYOiAIgwO31fJMlSzouPSDyx4iA998Nt4loM7tLOEWgMKbocKcgIioIARImCBaWAiCrrKlcH2vL7Ylz/D9VqneIhV5LIB9F77VaZ4vPuA2oRCAaygEKAfASBm9yYWNsZRoLCAEYASABKgMAAgIiLAgBEgUCBIgCIBohIIAGWiNcyyPZ8yTKukJcK2Mn8OmpQ6uNjsa8Zp1P07WUIioIARImBAaIAiBgGbJRw5Nm8BZQIg/QP722R1a7TmNh1PnxcjVqJYmKfSAiKggBEiYGDogCIMDt9XyTJUs6Lj0g8seIgPffDbeJaDO7SzhFoDCm6HCnICIqCAESJggWlgIgq6ypXB9ry+2Jc/w/Vap3iIVeSyAfRe+1WmeLz7gNqEQg"
                    },
                    {
                        "type": "ics23:simple",
                        "key": "b3JhY2xl",
                        "data": "Cv4BCgZvcmFjbGUSILWuAqi7QA1Y943STZKRzdhqI/oM/3sihXiy0WvbYyjtGgkIARgBIAEqAQAiJQgBEiEBXVOth1oJa51+x41GKsXabapY7uL7OJssOImGEiPT298iJwgBEgEBGiDAh+1I3Z5vtDdvCoqbjqdXtDqMS2OQUuNb9RLtMehqtiInCAESAQEaIDqK5Rpdb442lPE4nvtBIBBC+gc6duYo7pQ4EWrTGTZjIicIARIBARogxinvyHFSBjSggcRdbDRvGaaNqaNLpwZ36rA6wxPCgW8iJQgBEiEBPO/QdDCqwo6da1w04OySSpstZJxfeGYlKuwfwmlBFOI="
                    }
                ]
            },
            "height": "139",
            "codespace": ""
        }
    }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"Cv4BCgZvcmFjbGUSILWuAqi7QA1Y943STZKRzdhqI/oM/3sihXiy0WvbYyjtGgkIARgBIAEqAQAiJQgBEiEBXVOth1oJa51+x41GKsXabapY7uL7OJssOImGEiPT298iJwgBEgEBGiDAh+1I3Z5vtDdvCoqbjqdXtDqMS2OQUuNb9RLtMehqtiInCAESAQEaIDqK5Rpdb442lPE4nvtBIBBC+gc6duYo7pQ4EWrTGTZjIicIARIBARogxinvyHFSBjSggcRdbDRvGaaNqaNLpwZ36rA6wxPCgW8iJQgBEiEBPO/QdDCqwo6da1w04OySSpstZJxfeGYlKuwfwmlBFOI=",
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
					m.ParamsToRestakeStoresMerkleHash,
				),
				m.RollingseedToTransferStoresMerkleHash,
			),
			m.TssToUpgradeStoresMerkleHash,
		),
	)

	require.Equal(t, expectAppHash, apphash)
}
