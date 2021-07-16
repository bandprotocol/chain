package proof

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"
)

// BlockHeaderMerkleParts stores a group of hashes using for computing Tendermint's block
// header hash from app hash, and height.
//
// In Tendermint, a block header hash is the Merkle hash of a binary tree with 14 leaf nodes.
// Each node encodes a data piece of the blockchain. The notable data leaves are: [A] app_hash,
// [2] height. All data pieces are combined into one 32-byte hash to be signed
// by block validators. The structure of the Merkle tree is shown below.
//
//                                   [BlockHeader]
//                                /                \
//                   [3A]                                    [3B]
//                 /      \                                /      \
//         [2A]                [2B]                [2C]                [2D]
//        /    \              /    \              /    \              /    \
//    [1A]      [1B]      [1C]      [1D]      [1E]      [1F]        [C]    [D]
//    /  \      /  \      /  \      /  \      /  \      /  \
//  [0]  [1]  [2]  [3]  [4]  [5]  [6]  [7]  [8]  [9]  [A]  [B]
//
//  [0] - version               [1] - chain_id            [2] - height        [3] - time
//  [4] - last_block_id         [5] - last_commit_hash    [6] - data_hash     [7] - validators_hash
//  [8] - next_validators_hash  [9] - consensus_hash      [A] - app_hash      [B] - last_results_hash
//  [C] - evidence_hash         [D] - proposer_address
//
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [2], [3], and [A]. In fact, only
// [1A], [2B], [1E], [B], and [2D] are needed in order to compute [BlockHeader].
type BlockHeaderMerkleParts struct {
	VersionAndChainIdHash             tmbytes.HexBytes `json:"version_and_chain_id_hash"`
	Height                            uint64           `json:"height"`
	TimeSecond                        uint64           `json:"time_second"`
	TimeNanoSecond                    uint32           `json:"time_nano_second"`
	LastBlockIdAndOther               tmbytes.HexBytes `json:"last_block_id_and_other"`
	NextValidatorHashAndConsensusHash tmbytes.HexBytes `json:"next_validator_hash_and_consensus_hash"`
	LastResultsHash                   tmbytes.HexBytes `json:"last_results_hash"`
	EvidenceAndProposerHash           tmbytes.HexBytes `json:"evidence_and_proposer_hash"`
}

// BlockHeaderMerklePartsEthereum is an Ethereum version of BlockHeaderMerkleParts for solidity ABI-encoding.
type BlockHeaderMerklePartsEthereum struct {
	VersionAndChainIdHash             common.Hash
	Height                            uint64
	TimeSecond                        uint64
	TimeNanoSecond                    uint32
	LastBlockIdAndOther               common.Hash
	NextValidatorHashAndConsensusHash common.Hash
	LastResultsHash                   common.Hash
	EvidenceAndProposerHash           common.Hash
}

func (bp *BlockHeaderMerkleParts) encodeToEthFormat() BlockHeaderMerklePartsEthereum {
	return BlockHeaderMerklePartsEthereum{
		VersionAndChainIdHash:             common.BytesToHash(bp.VersionAndChainIdHash),
		Height:                            bp.Height,
		TimeSecond:                        bp.TimeSecond,
		TimeNanoSecond:                    bp.TimeNanoSecond,
		LastBlockIdAndOther:               common.BytesToHash(bp.LastBlockIdAndOther),
		NextValidatorHashAndConsensusHash: common.BytesToHash(bp.NextValidatorHashAndConsensusHash),
		LastResultsHash:                   common.BytesToHash(bp.LastResultsHash),
		EvidenceAndProposerHash:           common.BytesToHash(bp.EvidenceAndProposerHash),
	}
}

// GetBlockHeaderMerkleParts converts Tendermint block header struct into BlockHeaderMerkleParts for gas-optimized proof verification.
func GetBlockHeaderMerkleParts(block *types.Header) BlockHeaderMerkleParts {
	// based on https://github.com/tendermint/tendermint/blob/master/types/block.go#L448
	hbz, err := block.Version.Marshal()
	if err != nil {
		panic(err)
	}

	pbbi := block.LastBlockID.ToProto()
	bzbi, err := pbbi.Marshal()
	if err != nil {
		panic(err)
	}

	return BlockHeaderMerkleParts{
		VersionAndChainIdHash: merkle.HashFromByteSlices([][]byte{
			hbz,
			cdcEncode(block.ChainID),
		}),
		Height:         uint64(block.Height),
		TimeSecond:     uint64(block.Time.Unix()),
		TimeNanoSecond: uint32(block.Time.Nanosecond()),
		LastBlockIdAndOther: merkle.HashFromByteSlices([][]byte{
			bzbi,
			cdcEncode(block.LastCommitHash),
			cdcEncode(block.DataHash),
			cdcEncode(block.ValidatorsHash),
		}),
		NextValidatorHashAndConsensusHash: merkle.HashFromByteSlices([][]byte{
			cdcEncode(block.NextValidatorsHash),
			cdcEncode(block.ConsensusHash),
		}),
		LastResultsHash: merkle.HashFromByteSlices([][]byte{
			cdcEncode(block.LastResultsHash),
		}),
		EvidenceAndProposerHash: merkle.HashFromByteSlices([][]byte{
			cdcEncode(block.EvidenceHash),
			cdcEncode(block.ProposerAddress),
		}),
	}
}
