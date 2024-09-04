package proof

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	ics23 "github.com/confio/ics23/go"
	"github.com/stretchr/testify/require"
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

func getIAVLParentHash(path IAVLMerklePath, subtreeHash []byte) []byte {
	var lengthByte byte = 0x20

	// prefix of inner node
	preimage := convertVarIntToBytes(int64(path.SubtreeHeight))
	preimage = append(preimage, convertVarIntToBytes(int64(path.SubtreeSize))...)
	preimage = append(preimage, convertVarIntToBytes(int64(path.SubtreeVersion))...)

	if path.IsDataOnRight {
		preimage = append(preimage, lengthByte)
		preimage = append(preimage, path.SiblingHash...)

		preimage = append(preimage, lengthByte)
		preimage = append(preimage, subtreeHash...)
	} else {
		preimage = append(preimage, lengthByte)
		preimage = append(preimage, subtreeHash...)

		preimage = append(preimage, lengthByte)
		preimage = append(preimage, path.SiblingHash...)
	}

	return tmhash.Sum(preimage)
}

func TestGetMerklePaths(t *testing.T) {
	key := base64ToBytes("/wAAAAAAAAAB")
	data := base64ToBytes(
		"CocCCgn/AAAAAAAAAAESOxABGhMAAAABAAAAA0JUQwAAAAAAAYagIAEoATABOAFAk6mTiQZIl6mTiQZQAVoMAAAAAQAAAAEk7AeMGgwIARgBIAEqBAACsgMiKggBEiYCBLIDIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9ICIqCAESJgQIsgMgGEcQdQfV57TNmUHrb/4WlCZK80xoXBncR4vq3aJlpXggIioIARImBgyyAyDoCq5YHsAEI5hUxNkNgUjoXx+Q0HBKc2aP0tpE3AzqUyAiKwgBEicKIM6GAyB0G3lGJNvpTWfNae6o1Xqqdo0upy8fylYaS++2eDbp4SA=",
	)

	var iavlOps storetypes.CommitmentOp
	proof := &ics23.CommitmentProof{}
	err := proof.Unmarshal(data)
	require.Nil(t, err)

	iavlOps = storetypes.NewIavlCommitmentOp(key, proof)
	iavlEp := iavlOps.Proof.GetExist()
	require.NotNil(t, iavlEp)

	var expectOracleMerkleHash []byte
	expectOracleMerkleHash, err = iavlEp.Calculate()
	require.Nil(t, err)

	version := decodeIAVLLeafPrefix(iavlEp.Leaf.Prefix)
	value := iavlEp.Value

	leafNode := []byte{}
	leafNode = append(leafNode, convertVarIntToBytes(0)...)
	leafNode = append(leafNode, convertVarIntToBytes(1)...)
	leafNode = append(leafNode, convertVarIntToBytes(int64(version))...)
	leafNode = append(leafNode, uint8(len(key)))      // key length
	leafNode = append(leafNode, key...)               // key to result of request #1
	leafNode = append(leafNode, 32)                   // size of result hash must be 32
	leafNode = append(leafNode, tmhash.Sum(value)...) // value on this key is a result hash
	currentHash := tmhash.Sum(leafNode)

	paths := GetMerklePaths(iavlEp)
	for _, path := range paths {
		currentHash = getIAVLParentHash(path, currentHash)
	}
	require.Equal(t, expectOracleMerkleHash, currentHash)
}
