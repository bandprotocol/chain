package proof

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MultiStoreProof stores a compact of other Cosmos-SDK modules' storage hash in multistore to
// compute (in combination with oracle store hash) Tendermint's application state hash at a given block.
//                         ________________[AppHash]_______________
//                        /                                        \
//             _______[I9]______                          ________[I10]________
//            /                  \                       /                     \
//       __[I5]__             __[I6]__              __[I7]__               __[I8]__
//      /         \          /         \           /         \            /         \
//    [I1]       [I2]     [I3]        [I4]       [8]        [9]          [A]        [B]
//   /   \      /   \    /    \      /    \
// [0]   [1]  [2]   [3] [4]   [5]  [6]    [7]
// [0] - acc      [1] - distr   [2] - evidence  [3] - gov
// [4] - main     [5] - mint    [6] - oracle    [7] - params
// [8] - slashing [9] - staking [A] - supply    [D] - upgrade
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [6] In fact, only
// [7], [I3], [I5], and [I10] are needed in order to compute [AppHash].
type MultiStoreProof struct {
	OracleIAVLStateHash tmbytes.HexBytes `json:"oracle_iavl_state_hash"`
	MerklePaths         []IAVLMerklePath `json:"merkle_paths"`
}

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash common.Hash
	MerklePaths         []IAVLMerklePathEthereum
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	parsePaths := make([]IAVLMerklePathEthereum, len(m.MerklePaths))
	for i, path := range m.MerklePaths {
		parsePaths[i] = path.encodeToEthFormat()
	}
	return MultiStoreProofEthereum{
		OracleIAVLStateHash: common.BytesToHash(m.OracleIAVLStateHash),
		MerklePaths:         parsePaths,
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	paths := make([]IAVLMerklePath, 0)
	for _, step := range multiStoreEp.Path {
		if step.Hash != ics23.HashOp_SHA256 {
			// Currently tendermint is using SHA256 only, so we hardcode it for now.
			return MultiStoreProof{}
		}
		imp := IAVLMerklePath{
			step.Prefix,
			step.Suffix,
		}
		paths = append(paths, imp)
	}

	hash, err := multiStoreEp.Leaf.Apply(multiStoreEp.Key, multiStoreEp.Value)
	if err != nil {
		return MultiStoreProof{}
	}
	return MultiStoreProof{
		OracleIAVLStateHash: hash,
		MerklePaths:         paths,
	}
}
