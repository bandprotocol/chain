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
	jsonrpc: "2.0",
	id: -1,
	result: {
		response: {
			code: 0,
			log: "",
			info: "",
			index: "0",
			key: "/wAAAAAAAAAB",
			value: "EAEaEwAAAAEAAAADQlRDAAAAAAABhqAgASgBMAE4AUCTqZOJBkiXqZOJBlABWgwAAAABAAAAASTsB4w=",
			proofOps: {
				ops: [
					{
						type: "ics23:iavl",
						key: "/wAAAAAAAAAB",
						data: "CocCCgn/AAAAAAAAAAESOxABGhMAAAABAAAAA0JUQwAAAAAAAYagIAEoATABOAFAk6mTiQZIl6mTiQZQAVoMAAAAAQAAAAEk7AeMGgwIARgBIAEqBAACsgMiKggBEiYCBLIDIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9ICIqCAESJgQIsgMgGEcQdQfV57TNmUHrb/4WlCZK80xoXBncR4vq3aJlpXggIioIARImBgyyAyDoCq5YHsAEI5hUxNkNgUjoXx+Q0HBKc2aP0tpE3AzqUyAiKwgBEicKIM6GAyB0G3lGJNvpTWfNae6o1Xqqdo0upy8fylYaS++2eDbp4SA="
					},
					{
						type: "ics23:simple",
						key: "b3JhY2xl",
						data: "CvwBCgZvcmFjbGUSIOjifLtEu2VPZO7vRmeGitSGZ86yjj21xN96S0uH8MBLGgkIARgBIAEqAQAiJQgBEiEB+YFxZWKkneBuPcr7+2OIwpS6pPqdRXd+JXQKkvgc9l4iJQgBEiEBf9X1x8KSDBh2GFQpAc3FcXvoIE8kvoVugJAqG7BHN+QiJwgBEgEBGiD8ls/9MOW4l56mb50NocurFvaWaeiyofsuG+tFfJcm6CIlCAESIQFSSEY+ky0W99CS4mjA3th7I9Ow5xhW8cauKqkfbHEzICInCAESAQEaIMnIhJ7RJcx2gTKcTSe4Ox/IrPeoZcnR0d9XXMpW9I2+"
					}
				]
			},
			height: "24999",
			codespace: ""
		}
	}
}
*/

func TestGetMultiStoreProof(t *testing.T) {
	key := []byte("oracle")
	data := base64ToBytes("CvwBCgZvcmFjbGUSIOjifLtEu2VPZO7vRmeGitSGZ86yjj21xN96S0uH8MBLGgkIARgBIAEqAQAiJQgBEiEB+YFxZWKkneBuPcr7+2OIwpS6pPqdRXd+JXQKkvgc9l4iJQgBEiEBf9X1x8KSDBh2GFQpAc3FcXvoIE8kvoVugJAqG7BHN+QiJwgBEgEBGiD8ls/9MOW4l56mb50NocurFvaWaeiyofsuG+tFfJcm6CIlCAESIQFSSEY+ky0W99CS4mjA3th7I9Ow5xhW8cauKqkfbHEzICInCAESAQEaIMnIhJ7RJcx2gTKcTSe4Ox/IrPeoZcnR0d9XXMpW9I2+")

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
			m.AuthToFeeGrantStoresMerkleHash,
			innerHash(
				innerHash(
					m.GovToIbcCoreStoresMerkleHash,
					innerHash(
						m.MintStoreMerkleHash,
						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
					),
				),
				m.ParamsToTransferStoresMerkleHash,
			),
		),
		m.UpgradeStoreMerkleHash,
	)

	require.Equal(t, expectAppHash, apphash)
}
