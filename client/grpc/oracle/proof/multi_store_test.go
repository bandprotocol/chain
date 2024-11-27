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
                        "data": "EpADCgjAAAAAAAAAABK2AQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSEZQgOCllrsRWAEaDAgBGAEgASoEAAKYASIsCAESBQQG7hAgGiEgL9OKizGo0Z4tLP9KaLTbPfPi2jAZVEgXXK18Znu4V9EiKggBEiYGCu4QIGlHkf33ommIkdMfhPuQ14UG07Xp7lKmlq9ZNJSfVOlzICIqCAESJggY0GcgWmuKQaZpAh5otqBxJfSoJWrASawo/qPB1V5shwYbbgYgGsoBCgHwEgZvcmFjbGUaCwgBGAEgASoDAAICIiwIARIFAgTuECAaISAeDSwCB2vKaX7NoctkSGZ+6OCwUOgp6yg4Ye20FyJXSiIqCAESJgQG7hAgzL3GBeV7EMxT1yR/MzvVQdVSa8kOaTp4k0AkbuqN7XcgIioIARImBgruECBpR5H996JpiJHTH4T7kNeFBtO16e5SppavWTSUn1TpcyAiKggBEiYIGNBnIFprikGmaQIeaLagcSX0qCVqwEmsKP6jwdVebIcGG24GIA=="
                    },
                    {
                        "type": "ics23:simple",
                        "key": "b3JhY2xl",
                        "data": "Cv4BCgZvcmFjbGUSIKlGw34VS+Fki9oMhwxL5Augva71DvOvN6SD4WqsMjAjGgkIARgBIAEqAQAiJwgBEgEBGiCI0bT8lYEcR7qf3dcswRSJQN0ZgJdSo0ioy2uIsi/bZSIlCAESIQFjDp0uLX5qOZ6OLF4/FVCnrviBSSHovSU4TVjJiaS2myInCAESAQEaICjBhguC1q2OSs82GtyruVNh5egqJ15XXOcnOk6HMuJcIicIARIBARogaTJx7rOygthReqj7mtQImvchyp0oRDKbBMhhEN/zjpUiJQgBEiEB8dVwesb0WvvZh6/6bec9nDEIE+ms2MlXWKvAFKRT3fs="
                    }
                ]
            },
            "height": "6632",
            "codespace": ""
        }
    }
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes(
		"Cv4BCgZvcmFjbGUSIKlGw34VS+Fki9oMhwxL5Augva71DvOvN6SD4WqsMjAjGgkIARgBIAEqAQAiJwgBEgEBGiCI0bT8lYEcR7qf3dcswRSJQN0ZgJdSo0ioy2uIsi/bZSIlCAESIQFjDp0uLX5qOZ6OLF4/FVCnrviBSSHovSU4TVjJiaS2myInCAESAQEaICjBhguC1q2OSs82GtyruVNh5egqJ15XXOcnOk6HMuJcIicIARIBARogaTJx7rOygthReqj7mtQImvchyp0oRDKbBMhhEN/zjpUiJQgBEiEB8dVwesb0WvvZh6/6bec9nDEIE+ms2MlXWKvAFKRT3fs=",
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
			m.TransferToUpgradeStoresMerkleHash,
		),
	)

	require.Equal(t, expectAppHash, apphash)
}
