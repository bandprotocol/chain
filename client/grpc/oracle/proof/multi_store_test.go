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
                        "data": "EpADCgjAAAAAAAAAABK2AQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSEZQgOCllrsRWAEaDAgBGAEgASoEAAKYASIsCAESBQQG7hAgGiEgL9OKizGo0Z4tLP9KaLTbPfPi2jAZVEgXXK18Znu4V9EiKggBEiYGCu4QIGlHkf33ommIkdMfhPuQ14UG07Xp7lKmlq9ZNJSfVOlzICIqCAESJggY+mQgNsCXHqSI4Wc4nRUcu3B7JtZznzlsPzQs4gL/LaWWTCMgGsoBCgHwEgZvcmFjbGUaCwgBGAEgASoDAAICIiwIARIFAgTuECAaISAeDSwCB2vKaX7NoctkSGZ+6OCwUOgp6yg4Ye20FyJXSiIqCAESJgQG7hAgzL3GBeV7EMxT1yR/MzvVQdVSa8kOaTp4k0AkbuqN7XcgIioIARImBgruECBpR5H996JpiJHTH4T7kNeFBtO16e5SppavWTSUn1TpcyAiKggBEiYIGPpkIDbAlx6kiOFnOJ0VHLtweybWc585bD80LOIC/y2llkwjIA=="
                    },
                    {
                        "type": "ics23:simple",
                        "key": "b3JhY2xl",
                        "data": "Cv4BCgZvcmFjbGUSICxrji3gkPqdGI8I4w6Cm9w6v89awbV8uoSkKV14K0JpGgkIARgBIAEqAQAiJwgBEgEBGiCI0bT8lYEcR7qf3dcswRSJQN0ZgJdSo0ioy2uIsi/bZSIlCAESIQEEZvq0c6FbK1S9zuMOkLgw5Bxfpao3t2yexQtFgWPdESInCAESAQEaIEiSFz+L5H6FV+AQl38DlIcmMLN2Lyn60r59Izq+loBvIicIARIBARogzAP2Fagk8FbZUa26ZDPyT5yv2E6PD5KVQPkvi4fFMwMiJQgBEiEBHxefupAzs5mcxY3iwVwvEza8VywaWZiHj7j7LUDez9o="
                    }
                ]
            },
            "height": "6461",
            "codespace": ""
        }
    }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"Cv4BCgZvcmFjbGUSICxrji3gkPqdGI8I4w6Cm9w6v89awbV8uoSkKV14K0JpGgkIARgBIAEqAQAiJwgBEgEBGiCI0bT8lYEcR7qf3dcswRSJQN0ZgJdSo0ioy2uIsi/bZSIlCAESIQEEZvq0c6FbK1S9zuMOkLgw5Bxfpao3t2yexQtFgWPdESInCAESAQEaIEiSFz+L5H6FV+AQl38DlIcmMLN2Lyn60r59Izq+loBvIicIARIBARogzAP2Fagk8FbZUa26ZDPyT5yv2E6PD5KVQPkvi4fFMwMiJQgBEiEBHxefupAzs5mcxY3iwVwvEza8VywaWZiHj7j7LUDez9o=",
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
		m.AuthToIbcStoresMerkleHash,
		innerHash(
			innerHash(
				innerHash(
					m.IcahostToMintStoresMerkleHash,
					innerHash(
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
						m.ParamsStoreMerkleHash,
					),
				),
				m.RestakeToStakingStoresMerkleHash,
			),
			m.TransferToUpgradeStoreMerkleHash,
		),
	)

	require.Equal(t, expectAppHash, apphash)
}
