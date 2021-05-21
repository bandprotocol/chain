package proof

import (
	"encoding/binary"
	"math/big"

	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MerklePath represents a Merkle step to a leaf data node in an iAVL tree.
// TODO: Prefix & Suffix are currently dynamic-sized bytes, make them fixed
type IAVLMerklePath struct {
	IsDataOnRight  bool             `json:"is_data_on_right"`
	SubtreeHeight  uint8            `json:"subtree_height"`
	SubtreeSize    uint64           `json:"subtree_size"`
	SubtreeVersion uint64           `json:"subtree_version"`
	SiblingHash    tmbytes.HexBytes `json:"sibling_hash"`
}

// IAVLMerklePathEthereum is an Ethereum version of IAVLMerklePath for solidity ABI-encoding.
type IAVLMerklePathEthereum struct {
	IsDataOnRight  bool
	SubtreeHeight  uint8
	SubtreeSize    *big.Int
	SubtreeVersion *big.Int
	SiblingHash    common.Hash
}

func (merklePath *IAVLMerklePath) encodeToEthFormat() IAVLMerklePathEthereum {
	return IAVLMerklePathEthereum{
		merklePath.IsDataOnRight,
		merklePath.SubtreeHeight,
		big.NewInt(int64(merklePath.SubtreeSize)),
		big.NewInt(int64(merklePath.SubtreeVersion)),
		common.BytesToHash(merklePath.SiblingHash),
	}
}

func decodeIAVLLeafPrefix(prefix []byte) uint64 {
	// ref: https://github.com/cosmos/iavl/blob/master/proof_ics23.go#L96
	_, n1 := binary.Varint(prefix)
	_, n2 := binary.Varint(prefix[n1:])
	version, _ := binary.Varint(prefix[n1+n2:])
	return uint64(version)
}

// GetMerklePaths returns the list of MerklePath elements from the given iAVL proof.
func GetMerklePaths(iavlEp *ics23.ExistenceProof) []IAVLMerklePath {
	paths := make([]IAVLMerklePath, 0)
	for _, step := range iavlEp.Path {
		if step.Hash != ics23.HashOp_SHA256 {
			// Tendermint v0.34.9 is using SHA256 only.
			panic("Expect HashOp_SHA256")
		}
		imp := IAVLMerklePath{}

		// decode IAVL inner prefix
		// ref: https://github.com/cosmos/iavl/blob/master/proof_ics23.go#L96
		subtreeHeight, n1 := binary.Varint(step.Prefix)
		subtreeSize, n2 := binary.Varint(step.Prefix[n1:])
		subtreeVersion, n3 := binary.Varint(step.Prefix[n1+n2:])

		imp.SubtreeHeight = uint8(subtreeHeight)
		imp.SubtreeSize = uint64(subtreeSize)
		imp.SubtreeVersion = uint64(subtreeVersion)

		prefixLength := n1 + n2 + n3 + 1
		if prefixLength != len(step.Prefix) {
			imp.IsDataOnRight = true
			imp.SiblingHash = step.Prefix[prefixLength : len(step.Prefix)-1]
		} else {
			imp.IsDataOnRight = false
			imp.SiblingHash = step.Suffix[1:]
		}
		paths = append(paths, imp)
	}
	return paths
}

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
