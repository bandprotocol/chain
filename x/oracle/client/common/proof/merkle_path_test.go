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
	"code": 0,
	"log": "",
	"info": "",
	"index": "0",
	"key": "/wAAAAAAAAAB",
	"value": "AAAAAAAAAAAAAAABAAAADwAAAANCVEMAAAAAAAAD6AAAAAAAAAABAAAAAAAAAAEAAAAAAAAAAQAAAAAAAAABAAAAAGCKeEgAAAAAYIp4SgAAAAEAAAAIAAAAAAM4YF8=",
	"proofOps": {
		"ops": [
			{
			"type": "ics23:iavl",
			"key": "/wAAAAAAAAAB",
			"data": "CtcCCgn/AAAAAAAAAAESXwAAAAAAAAAAAAAAAQAAAA8AAAADQlRDAAAAAAAAA+gAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAEAAAAAAAAAAQAAAABginhIAAAAAGCKeEoAAAABAAAACAAAAAADOGBfGgwIARgBIAEqBAAC7gQiKggBEiYCBO4EIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9ICIqCAESJgQI7gQg7Gs6hyT0aoCv6lr/Gy++Ks8FMrTz99Ltaua4F3EzUvIgIioIARImBgzuBCA3Wh3tcSTfglD/FmGkQdAR6MBiwPIjkQ2fXYOH9fOmMyAiKggBEiYIFO4EIOsThN9b1u/lW1m2SgpRNYiSu55ZsYwhCe6uFmPhWroEICIrCAESJwoqpP0mIAeuDTGG5ahwRTqLDI/XkTi5IiiX8jYD89TY9Hj0oVjXIA=="
			},
			{
			"type": "ics23:simple",
			"key": "b3JhY2xl",
			"data": "CtUBCgZvcmFjbGUSIGCpOg2TAwJdW9s4uMmvU5jS9PxFxDYeUSuyqB9PCrDcGgkIARgBIAEqAQAiJQgBEiEBUDdkoTaAKkGcRCOX8uLZogDz1MO9rGC6labK+12nqOEiJwgBEgEBGiDnC0likzrQRzkAC/7rjekGhzgDxYb1rTD0fcNY46A4NCInCAESAQEaIPidKKjCBeXa/TTJSgg/XzaXyeNLvP+Sap/tOEh0tJ9uIiUIARIhAd8npY8q61YdtX9zAmp7jE4juFGvzZcw4wo1tKSdFCiw"
			}
		]
	},
	"height": "319314",
	"codespace": ""
}
*/

func TestGetMerklePaths(t *testing.T) {
	key := base64ToBytes("/wAAAAAAAAAB")
	data := base64ToBytes("CtcCCgn/AAAAAAAAAAESXwAAAAAAAAAAAAAAAQAAAA8AAAADQlRDAAAAAAAAA+gAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAEAAAAAAAAAAQAAAABginhIAAAAAGCKeEoAAAABAAAACAAAAAADOGBfGgwIARgBIAEqBAAC7gQiKggBEiYCBO4EIOtzm7IvSLfzBTqQuiuk/gf6smLK34ZkSJVlxQ/1Bbi9ICIqCAESJgQI7gQg7Gs6hyT0aoCv6lr/Gy++Ks8FMrTz99Ltaua4F3EzUvIgIioIARImBgzuBCA3Wh3tcSTfglD/FmGkQdAR6MBiwPIjkQ2fXYOH9fOmMyAiKggBEiYIFO4EIOsThN9b1u/lW1m2SgpRNYiSu55ZsYwhCe6uFmPhWroEICIrCAESJwoqpP0mIAeuDTGG5ahwRTqLDI/XkTi5IiiX8jYD89TY9Hj0oVjXIA==")

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

	value := iavlEp.Value

	leafNode := []byte{}
	leafNode = append(leafNode, iavlEp.Leaf.Prefix...) // leaf prefix
	leafNode = append(leafNode, uint8(len(key)))       // key length
	leafNode = append(leafNode, key...)                // key to result of request #1
	leafNode = append(leafNode, 32)                    // size of result hash must be 32
	leafNode = append(leafNode, tmhash.Sum(value)...)  // value on this key is a result hash
	currentHash := tmhash.Sum(leafNode)

	paths := GetMerklePaths(iavlEp)
	for _, path := range paths {
		currentHash = getParentHash(path, currentHash)
	}
	require.Equal(t, expectOracleMerkleHash, currentHash)
}
