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
      	"hash": "EF40EA6FC7CACE83B8496B5820D85076505B1F0FF5995F22EAF5A66A5057E168",
      	"parts": {
        	"total": 1,
        	"hash": "64F72682F614F027E37900A985E01251213FCDCA048FF2F3C42A42B90E6AD8F8"
      	}
    },
    "block": {
      	"header": {
        	"version": {
          	"block": "11"
			},
			"chain_id": "band-laozi-testnet1",
			"height": "180355",
			"time": "2021-05-19T08:20:41.556046438Z",
			"last_block_id": {
				"hash": "29A827A2F05C3E9C6FA63096E5D02E3E19E05854151CA7ADDDC93D64171FF692",
				"parts": {
					"total": 1,
					"hash": "C0E491BD75BCD1773FEB76332488ADDAE5F64B2D9232E34C481981AF4701F273"
				}
			},
			"last_commit_hash": "C11805F28626CDE072923DD24FF29BA6AB788C6520AA00A5C20E5B512CC2EF3F",
			"data_hash": "E05A3E896C1FBB3DD873CDBCB00D026302BAA813C9ADBA8A8A860AA3C653AE84",
			"validators_hash": "372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C",
			"next_validators_hash": "372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C",
			"consensus_hash": "BED75E0A0CDB709FBA26EA5D58D4207C32F9DBF96634F8B0F07D01FC06132AAC",
			"app_hash": "C6122B5FB927E4EEC152862FE46ED62834ED749DD056113D7A1F053541586F17",
			"last_results_hash": "329482B1DD2307DF32C2FC6E044451D62FCA4F51C6F37BC5A6195B9AB8FE8195",
			"evidence_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			"proposer_address": "7C4E472143CFCF54E5304AD43F97D13C9A962AC0"
		}
	}
}
*/
func TestBlockHeaderMerkleParts(t *testing.T) {
	// Copy block header Merkle Part here
	header := types.Header{
		Version: tmversion.Consensus{Block: 11},
		ChainID: "band-laozi-testnet1",
		Height:  180355,
		Time:    parseTime("2021-05-19T08:20:41.556046438Z"),
		LastBlockID: types.BlockID{
			Hash: hexToBytes("29A827A2F05C3E9C6FA63096E5D02E3E19E05854151CA7ADDDC93D64171FF692"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("C0E491BD75BCD1773FEB76332488ADDAE5F64B2D9232E34C481981AF4701F273"),
			},
		},
		LastCommitHash:     hexToBytes("C11805F28626CDE072923DD24FF29BA6AB788C6520AA00A5C20E5B512CC2EF3F"),
		DataHash:           hexToBytes("E05A3E896C1FBB3DD873CDBCB00D026302BAA813C9ADBA8A8A860AA3C653AE84"),
		ValidatorsHash:     hexToBytes("372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C"),
		NextValidatorsHash: hexToBytes("372352B297752AF3687FE8755313FBFDE89515D9EB7761BD2E3C8EEBE7FBA63C"),
		ConsensusHash:      hexToBytes("BED75E0A0CDB709FBA26EA5D58D4207C32F9DBF96634F8B0F07D01FC06132AAC"),
		AppHash:            hexToBytes("C6122B5FB927E4EEC152862FE46ED62834ED749DD056113D7A1F053541586F17"),
		LastResultsHash:    hexToBytes("329482B1DD2307DF32C2FC6E044451D62FCA4F51C6F37BC5A6195B9AB8FE8195"),
		EvidenceHash:       hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		ProposerAddress:    hexToBytes("7C4E472143CFCF54E5304AD43F97D13C9A962AC0"),
	}
	blockMerkleParts := GetBlockHeaderMerkleParts(&header)
	expectBlockHash := hexToBytes("EF40EA6FC7CACE83B8496B5820D85076505B1F0FF5995F22EAF5A66A5057E168")
	appHash := tmbytes.HexBytes(hexToBytes("C6122B5FB927E4EEC152862FE46ED62834ED749DD056113D7A1F053541586F17"))

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
