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
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "response": {
      "code": 0,
      "log": "",
      "info": "",
      "index": "0",
      "key": "/wAAAAAAAAAB",
      "value": "AAAACmZyb21fYmFuZGQAAAAAAAAAIQAAAAgAAAAAAAAAZAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAAQAAAAAAAAAQAAAAAGCdRR4AAAAAYJ1FJgAAAAEAAAAIAAAAAAACyIk=",
      "proofOps": {
        "ops": [
          {
            "type": "ics23:iavl",
            "key": "/wAAAAAAAAAB",
            "data": "CvMGCgn/AAAAAAAAAAESYgAAAApmcm9tX2JhbmRkAAAAAAAAACEAAAAIAAAAAAAAAGQAAAAAAAAAEAAAAAAAAAAQAAAAAAAAAAEAAAAAAAAAEAAAAABgnUUeAAAAAGCdRSYAAAABAAAACAAAAAAAAsiJGgwIARgBIAEqBAACkAYiLAgBEgUCBOgGIBohIGdj7fQsDXo3ZejNm5cK4OINxtPPXfDcY8rSyF+vxqgDIioIARImBAjuBiCS8zYBRmdp1iZwpYdxyPjyaV5xQrOFIZfdPKaCW4o7JiAiLAgBEgUGEJAJIBohIFLEslBD/3YNtK4/NB6DCQgATR58O733JLxx3CSqaFE0Ii0IARIGCCCyqQ0gGiEgwJJO/Pr3fk/2Xp8k7QxDx7u7sHDMERxKWNorZrEYnnQiLQgBEgYKNpSsDSAaISBWmsxce+2sB+RRqtS0y0jfglzQN8kcfAz2A6rw1xVa0yItCAESBgxoxsANIBohIIlpQWHoQmHVKLmF8YScJw4GgcHeVj1UL/h0zkqMqkPPIi4IARIHDtoB+soNIBohIDkgQjAtnMeQYgCbYUy1c5yCJ+5HorG9Q27XoBtBMBhmIi4IARIHEMwDlNINIBohIGG8Ftxs3vXSA2bbB+sp0yjc+g1lQeewmmcV8t6dCXkUIi4IARIHEo4Hzu0NIBohIBoSS3Q0SHgfAQm1WZLXEorbT78bknoBDFiS0RGqvwktIi4IARIHFMQO+v8NIBohIJYOQzto4+xF6CaLNEjqiy3pQQkbRp88BIQAsAdJ6Tf2Ii4IARIHFsgc0KkOIBohIE6kjn74Dkv6lr193sHqaXKQ1BfP7wzhMB5pR9gNwKzDIiwIARIoGLAphoIWILhp1b7zuFYz6LqSZ/VHs+2O+B7a5Z562Qut7p9gx3t5ICIuCAESBxrsZIaCFiAaISBTap5a2qYuQSOnlGwHDEZZ5gHHZPHdg0SuMt3S8MDroyIvCAESCBzG3AGGghYgGiEgHtxX9LH3jy1lUMKuGY5xiQQDeti4z5mjUUXfdifmNmAiLwgBEggevMgDhoIWIBohIIikNBin0veShDGow/Is2j75y5fbYW95jXzx89Rz6GhOIi8IARIIIP6eBoaCFiAaISC57L3ZFn7voKNCFn9Fb/a+b6nFa2qyZcQs7NKDf7p/Yw=="
          },
          {
            "type": "ics23:simple",
            "key": "b3JhY2xl",
            "data": "CtUBCgZvcmFjbGUSIJj83HwI9IC+eoJooHuGNTM9kChH7A6lYG8z1Douk2wOGgkIARgBIAEqAQAiJQgBEiEBrn8EGLzowJ0sM7mBpuomG6Mwx12I3BY3pFK8xlxa6MEiJwgBEgEBGiDgAE8rLdq18Z4gJ/jN5svn/CoLe/ou9Iu2FPhZERPL8CInCAESAQEaIO8Ux+H17c0lq2FuOUtu2JYfZu0rw2Nge1D887onYMb4IiUIARIhAX+pMhUpuZRYyJ9LGxYmssLATEHrDkf8vS+6fqeLnWXX"
          }
        ]
      },
      "height": "180355",
      "codespace": ""
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
	data := base64ToBytes("CvMGCgn/AAAAAAAAAAESYgAAAApmcm9tX2JhbmRkAAAAAAAAACEAAAAIAAAAAAAAAGQAAAAAAAAAEAAAAAAAAAAQAAAAAAAAAAEAAAAAAAAAEAAAAABgnUUeAAAAAGCdRSYAAAABAAAACAAAAAAAAsiJGgwIARgBIAEqBAACkAYiLAgBEgUCBOgGIBohIGdj7fQsDXo3ZejNm5cK4OINxtPPXfDcY8rSyF+vxqgDIioIARImBAjuBiCS8zYBRmdp1iZwpYdxyPjyaV5xQrOFIZfdPKaCW4o7JiAiLAgBEgUGEJAJIBohIFLEslBD/3YNtK4/NB6DCQgATR58O733JLxx3CSqaFE0Ii0IARIGCCCyqQ0gGiEgwJJO/Pr3fk/2Xp8k7QxDx7u7sHDMERxKWNorZrEYnnQiLQgBEgYKNpSsDSAaISBWmsxce+2sB+RRqtS0y0jfglzQN8kcfAz2A6rw1xVa0yItCAESBgxoxsANIBohIIlpQWHoQmHVKLmF8YScJw4GgcHeVj1UL/h0zkqMqkPPIi4IARIHDtoB+soNIBohIDkgQjAtnMeQYgCbYUy1c5yCJ+5HorG9Q27XoBtBMBhmIi4IARIHEMwDlNINIBohIGG8Ftxs3vXSA2bbB+sp0yjc+g1lQeewmmcV8t6dCXkUIi4IARIHEo4Hzu0NIBohIBoSS3Q0SHgfAQm1WZLXEorbT78bknoBDFiS0RGqvwktIi4IARIHFMQO+v8NIBohIJYOQzto4+xF6CaLNEjqiy3pQQkbRp88BIQAsAdJ6Tf2Ii4IARIHFsgc0KkOIBohIE6kjn74Dkv6lr193sHqaXKQ1BfP7wzhMB5pR9gNwKzDIiwIARIoGLAphoIWILhp1b7zuFYz6LqSZ/VHs+2O+B7a5Z562Qut7p9gx3t5ICIuCAESBxrsZIaCFiAaISBTap5a2qYuQSOnlGwHDEZZ5gHHZPHdg0SuMt3S8MDroyIvCAESCBzG3AGGghYgGiEgHtxX9LH3jy1lUMKuGY5xiQQDeti4z5mjUUXfdifmNmAiLwgBEggevMgDhoIWIBohIIikNBin0veShDGow/Is2j75y5fbYW95jXzx89Rz6GhOIi8IARIIIP6eBoaCFiAaISC57L3ZFn7voKNCFn9Fb/a+b6nFa2qyZcQs7NKDf7p/Yw==")

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
