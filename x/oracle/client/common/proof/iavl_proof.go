package proof

import (
	ics23 "github.com/confio/ics23/go"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// IAVLMerklePath represents a Merkle step to a leaf data node in an iAVL tree.
// TODO: Prefix & Suffix are currently dynamic-sized bytes, make them fixed
type IAVLMerklePath struct {
	Prefix tmbytes.HexBytes `json:"prefix"`
	Suffix tmbytes.HexBytes `json:"suffix"`
}

// IAVLMerklePathEthereum is an Ethereum version of IAVLMerklePath for solidity ABI-encoding.
type IAVLMerklePathEthereum struct {
	Prefix []byte
	Suffix []byte
}

func (merklePath *IAVLMerklePath) encodeToEthFormat() IAVLMerklePathEthereum {
	return IAVLMerklePathEthereum{
		Prefix: merklePath.Prefix,
		Suffix: merklePath.Suffix,
	}
}

// GetIAVLMerklePaths returns the list of IAVLMerklePath elements from the given iAVL proof.
func GetIAVLMerklePaths(iavlEp *ics23.ExistenceProof) []IAVLMerklePath {
	paths := make([]IAVLMerklePath, 0)
	for _, step := range iavlEp.Path {
		if step.Hash != ics23.HashOp_SHA256 {
			// Tendermint v0.34.9 is using SHA256 only.
			panic("Expect HashOp_SHA256")
		}
		imp := IAVLMerklePath{
			step.Prefix,
			step.Suffix,
		}
		paths = append(paths, imp)
	}
	return paths
}
