package proof

var relayFormat = []byte(`
[
  {
    "components": [
      {
        "internalType": "bytes32",
        "name": "oracleIAVLStateHash",
        "type": "bytes32"
      },
      {
        "components": [
          {
            "internalType": "bytes",
            "name": "prefix",
            "type": "bytes"
          },
          {
            "internalType": "bytes",
            "name": "suffix",
            "type": "bytes"
          }
        ],
        "internalType": "struct IAVLMerklePath.Data[]",
        "name": "merklePaths",
        "type": "tuple[]"
      }
    ],
    "internalType": "struct MultiStore.Data",
    "name": "_multiStore",
    "type": "tuple"
  },
  {
    "components": [
      {
        "internalType": "bytes32",
        "name": "versionAndChainIDHash",
        "type": "bytes32"
      },
      {
        "internalType": "uint64",
        "name": "height",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "timeSecond",
        "type": "uint64"
      },
      {
        "internalType": "uint32",
        "name": "timeNanoSecond",
        "type": "uint32"
      },
      {
        "internalType": "bytes32",
        "name": "lastBlockIDAndOther",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "nextValidatorHashAndConsensusHash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "lastResultsHash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "evidenceAndProposerHash",
        "type": "bytes32"
      }
    ],
    "internalType": "struct BlockHeaderMerkleParts.Data",
    "name": "_merkleParts",
    "type": "tuple"
  },
  {
    "components": [
      {
        "internalType": "bytes32",
        "name": "r",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "s",
        "type": "bytes32"
      },
      {
        "internalType": "uint8",
        "name": "v",
        "type": "uint8"
      },
      {
        "internalType": "bytes",
        "name": "signedDataPrefix",
        "type": "bytes"
      },
      {
        "internalType": "bytes",
        "name": "signedDataSuffix",
        "type": "bytes"
      }
    ],
    "internalType": "struct TMSignature.Data[]",
    "name": "_signatures",
    "type": "tuple[]"
  }
]
`)

var verifyFormat = []byte(`
[
  {
    "internalType": "uint256",
    "name": "_blockHeight",
    "type": "uint256"
  },
  {
    "components": [
      {
        "internalType": "string",
        "name": "clientId",
        "type": "string"
      },
      {
        "internalType": "uint64",
        "name": "oracleScriptId",
        "type": "uint64"
      },
      {
        "internalType": "bytes",
        "name": "callData",
        "type": "bytes"
      },
      {
        "internalType": "uint64",
        "name": "askCount",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "minCount",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "requestId",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "ansCount",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "requestTime",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "resolveTime",
        "type": "uint64"
      },
      {
        "internalType": "uint8",
        "name": "resolveStatus",
        "type": "uint8"
      },
      {
        "internalType": "bytes",
        "name": "result",
        "type": "bytes"
      }
    ],
    "internalType": "struct IBridge.Result",
    "name": "_result",
    "type": "tuple"
  },
  {
    "internalType": "bytes",
    "name": "_prefix",
    "type": "bytes"
  },
  {
    "components": [
      {
        "internalType": "bytes",
        "name": "prefix",
        "type": "bytes"
      },
      {
        "internalType": "bytes",
        "name": "suffix",
        "type": "bytes"
      }
    ],
    "internalType": "struct IAVLMerklePath.Data[]",
    "name": "_merklePaths",
    "type": "tuple[]"
  }
]
`)

var verifyCountFormat = []byte(`
[
  {
    "internalType": "uint256",
    "name": "_blockHeight",
    "type": "uint256"
  },
  {
    "internalType": "uint256",
    "name": "_count",
    "type": "uint256"
  },
  {
    "internalType": "bytes",
    "name": "_prefix",
    "type": "bytes"
  },
  {
    "components": [
      {
        "internalType": "bytes",
        "name": "prefix",
        "type": "bytes"
      },
      {
        "internalType": "bytes",
        "name": "suffix",
        "type": "bytes"
      }
    ],
    "internalType": "struct IAVLMerklePath.Data[]",
    "name": "_merklePaths",
    "type": "tuple[]"
  }
]
`)
