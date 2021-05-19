package proof

import (
	ics23 "github.com/confio/ics23/go"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MerklePath represents a Merkle step to a leaf data node in an iAVL tree.
// TODO: Prefix & Suffix are currently dynamic-sized bytes, make them fixed
type MerklePath struct {
	Prefix tmbytes.HexBytes `json:"prefix"`
	Suffix tmbytes.HexBytes `json:"suffix"`
}

// MerklePathEthereum is an Ethereum version of MerklePath for solidity ABI-encoding.
type MerklePathEthereum struct {
	Prefix []byte
	Suffix []byte
}

func (merklePath *MerklePath) encodeToEthFormat() MerklePathEthereum {
	return MerklePathEthereum{
		Prefix: merklePath.Prefix,
		Suffix: merklePath.Suffix,
	}
}

// GetMerklePaths returns the list of MerklePath elements from the given iAVL proof.
func GetMerklePaths(iavlEp *ics23.ExistenceProof) []MerklePath {
	paths := make([]MerklePath, 0)
	for _, step := range iavlEp.Path {
		if step.Hash != ics23.HashOp_SHA256 {
			// Tendermint v0.34.9 is using SHA256 only.
			panic("Expect HashOp_SHA256")
		}
		imp := MerklePath{
			step.Prefix,
			step.Suffix,
		}
		paths = append(paths, imp)
	}
	return paths
}
