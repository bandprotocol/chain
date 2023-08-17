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
                        "data": "EsMDCgjAAAAAAAAAABLpAQoVBQ1eS8RW3fDH5wAMd46Ih4w7Tw0SEhAIARIMCPLi8Z8GEMDCo+ABGgwIARgBIAEqBAAC+CUiKggBEiYCBPglIMOXtvsiuUKm461Mtf/w7SDhGVqba09R5L1VU2Y2HGeOICIsCAESBQQI5CYgGiEgubs5NRGOw8xaesuOPr7hPvlY1LNxamfQ5qwMMKn8l80iKggBEiYGDOQmIAqGoK3PW8B0zAWNvdrvKtLMf9f2nUvJ/8HUfbG+mLCDICIqCAESJgga+isgwVjErYVeGYGfWDSJS9zW6U1GPMLdybMMGp3k+lT5x+ggGsoBCgHwEgZvcmFjbGUaCwgBGAEgASoDAAICIiwIARIFAgTkJiAaISCUHaTOnLkieD5s4z9auzNnWGOptSWVA3As2DTxLWzmZyIqCAESJgQI5CYg5zZMcUYFSwmSAmz9ub37BrR2v3mWJNOfELL/z1YSFJggIioIARImBgzkJiAKhqCtz1vAdMwFjb3a7yrSzH/X9p1Lyf/B1H2xvpiwgyAiKggBEiYIGvorIMFYxK2FXhmBn1g0iUvc1ulNRjzC3cmzDBqd5PpU+cfoIA=="
                    },
                    {
                        "type": "ics23:simple",
                        "key": "b3JhY2xl",
                        "data": "CvwBCgZvcmFjbGUSIIM4YL+a7OkzwrQJ6EwUJEt+FIGnboCro+OQiGZs9UAFGgkIARgBIAEqAQAiJQgBEiEBCdlpWuQ2tdx5sWp9VcTA+Y2NRuM1L8YtFLgJeZsaDxUiJwgBEgEBGiCwlrc6YwUsF2ydIB3k9PiBO3bBHhoP0IHOJtO2nDIVPyIlCAESIQFjiuPaA6L1DtVd8DXecpB+lu/MQBzkzCZ6M7j3fc3/8iIlCAESIQGDasyPrY+cK/pMuiQihLx69Ek6gZUDJ+b6v80jTKvZfiInCAESAQEaIP3tTGQiclpWi3Qerkck9TntQJo1rBxSoN6oEp1iFuNJ"
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
		"CvwBCgZvcmFjbGUSIIM4YL+a7OkzwrQJ6EwUJEt+FIGnboCro+OQiGZs9UAFGgkIARgBIAEqAQAiJQgBEiEBCdlpWuQ2tdx5sWp9VcTA+Y2NRuM1L8YtFLgJeZsaDxUiJwgBEgEBGiCwlrc6YwUsF2ydIB3k9PiBO3bBHhoP0IHOJtO2nDIVPyIlCAESIQFjiuPaA6L1DtVd8DXecpB+lu/MQBzkzCZ6M7j3fc3/8iIlCAESIQGDasyPrY+cK/pMuiQihLx69Ek6gZUDJ+b6v80jTKvZfiInCAESAQEaIP3tTGQiclpWi3Qerkck9TntQJo1rBxSoN6oEp1iFuNJ",
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
				m.GovToIcahostStoresMerkleHash,
				innerHash(
					innerHash(
						m.MintStoreMerkleHash,
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
					),
					m.ParamsToSlashingStoresMerkleHash,
				),
			),
		),
		m.StakingToUpgradeStoresMerkleHash,
	)

	require.Equal(t, expectAppHash, apphash)
}
