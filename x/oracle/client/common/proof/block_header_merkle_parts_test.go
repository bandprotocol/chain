package proof

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"
)

/*
{
	"block_id": {
      	"hash": "8C36C3D12A378BD7E4E8F26BDECCA68B48390240DA456EE9C3292B6E36756AC4",
		"parts": {
			"total": 1,
			"hash": "44551F853D916A7C630C0C210C921BAC7D05CE0C249DFC6088C0274F05841827"
		}
    },
    "block": {
      	"header": {
        	"version": {
				"block": "11"
			},
			"chain_id": "band-laozi-testnet1",
			"height": "180356",
			"time": "2021-05-19T08:20:43.922160838Z",
			"last_block_id": {
			"hash": "EF40EA6FC7CACE83B8496B5820D85076505B1F0FF5995F22EAF5A66A5057E168",
			"parts": {
				"total": 1,
				"hash": "64F72682F614F027E37900A985E01251213FCDCA048FF2F3C42A42B90E6AD8F8"
			}
			},
			"last_commit_hash": "40D61BC067EBC47C80DACF936616623274956FB3DB7A96E1717AD45F7A689DAC",
			"data_hash": "204EA8CEEAAE6D3E7C2DAC7B805049D241B7DB32252820567FDEFF6A97866BE8",
			"validators_hash": "372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C",
			"next_validators_hash": "372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C",
			"consensus_hash": "BED75E0A0CDB709FBA26EA5D58D4207C32F9DBF96634F8B0F07D01FC06132AAC",
			"app_hash": "E500B3DD21816EE04BE5E77271EC0D8286B8AFF81EF96344FED74B52992E6D23",
			"last_results_hash": "5E9B2DDD2AC52423AA2D0B04172EB4C464EFDCD2A00423D58FB71358E8BAFA18",
			"evidence_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			"proposer_address": "DC2E3CF55B246C881C1036C8D9F24BC23BA84AD3"
		}
	}
}
*/
func TestBlockHeaderMerkleParts(t *testing.T) {
	// Copy block header Merkle Part here
	header := types.Header{
		Version: tmversion.Consensus{Block: 11},
		ChainID: "band-laozi-testnet1",
		Height:  180356,
		Time:    parseTime("2021-05-19T08:20:43.922160838Z"),
		LastBlockID: types.BlockID{
			Hash: hexToBytes("EF40EA6FC7CACE83B8496B5820D85076505B1F0FF5995F22EAF5A66A5057E168"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("64F72682F614F027E37900A985E01251213FCDCA048FF2F3C42A42B90E6AD8F8"),
			},
		},
		LastCommitHash:     hexToBytes("40D61BC067EBC47C80DACF936616623274956FB3DB7A96E1717AD45F7A689DAC"),
		DataHash:           hexToBytes("204EA8CEEAAE6D3E7C2DAC7B805049D241B7DB32252820567FDEFF6A97866BE8"),
		ValidatorsHash:     hexToBytes("372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C"),
		NextValidatorsHash: hexToBytes("372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C"),
		ConsensusHash:      hexToBytes("BED75E0A0CDB709FBA26EA5D58D4207C32F9DBF96634F8B0F07D01FC06132AAC"),
		AppHash:            hexToBytes("E500B3DD21816EE04BE5E77271EC0D8286B8AFF81EF96344FED74B52992E6D23"),
		LastResultsHash:    hexToBytes("5E9B2DDD2AC52423AA2D0B04172EB4C464EFDCD2A00423D58FB71358E8BAFA18"),
		EvidenceHash:       hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		ProposerAddress:    hexToBytes("DC2E3CF55B246C881C1036C8D9F24BC23BA84AD3"),
	}
	blockMerkleParts := GetBlockHeaderMerkleParts(&header)
	expectBlockHash := hexToBytes("8C36C3D12A378BD7E4E8F26BDECCA68B48390240DA456EE9C3292B6E36756AC4")
	appHash := tmbytes.HexBytes(hexToBytes("E500B3DD21816EE04BE5E77271EC0D8286B8AFF81EF96344FED74B52992E6D23"))

	// Verify code
	blockHash := innerHash(
		innerHash(
			innerHash(
				blockMerkleParts.VersionAndChainIdHash,
				innerHash(
					leafHash(cdcEncode(header.Height)),
					leafHash(encodeTime(header.Time)),
				),
			),
			blockMerkleParts.LastBlockIdAndOther,
		),
		innerHash(
			innerHash(
				blockMerkleParts.NextValidatorHashAndConsensusHash,
				innerHash(
					leafHash(cdcEncode(appHash)),
					blockMerkleParts.LastResultsHash,
				),
			),
			blockMerkleParts.EvidenceAndProposerHash,
		),
	)
	require.Equal(t, expectBlockHash, blockHash)
}
