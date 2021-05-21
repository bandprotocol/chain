package proof

var relayFormat = []byte(`
[
  {
    "components": [
      {
        "internalType": "bytes32",
        "name": "authToIbcTransferStoresMerkleHash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "mintStoreMerkleHash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "oracleIAVLStateHash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "paramsToSlashStoresMerkleHash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes32",
        "name": "stakingToUpgradeStoresMerkleHash",
        "type": "bytes32"
      }
    ],
    "internalType": "struct MultiStore.Data",
    "name": "multiStore",
    "type": "tuple"
  },
  {
    "components": [
      {
        "internalType": "bytes32",
        "name": "versionAndChainIdHash",
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
        "name": "lastBlockIdAndOther",
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
    "name": "merkleParts",
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
    "name": "signatures",
    "type": "tuple[]"
  }
]
`)

var verifyFormat = []byte(`
[
  {
    "internalType": "uint256",
    "name": "blockHeight",
    "type": "uint256"
  },
  {
    "components": [
      {
        "internalType": "string",
        "name": "clientID",
        "type": "string"
      },
      {
        "internalType": "uint64",
        "name": "oracleScriptID",
        "type": "uint64"
      },
      {
        "internalType": "bytes",
        "name": "params",
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
        "name": "requestID",
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
    "name": "result",
    "type": "tuple"
  },
  {
    "internalType": "uint256",
    "name": "version",
    "type": "uint256"
  },
  {
    "components": [
      {
        "internalType": "bool",
        "name": "isDataOnRight",
        "type": "bool"
      },
      {
        "internalType": "uint8",
        "name": "subtreeHeight",
        "type": "uint8"
      },
      {
        "internalType": "uint256",
        "name": "subtreeSize",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "subtreeVersion",
        "type": "uint256"
      },
      {
        "internalType": "bytes32",
        "name": "siblingHash",
        "type": "bytes32"
      }
    ],
    "internalType": "struct IAVLMerklePath.Data[]",
    "name": "merklePaths",
    "type": "tuple[]"
  }
]
`)

var verifyCountFormat = []byte(`
[
  {
    "internalType": "uint256",
    "name": "blockHeight",
    "type": "uint256"
  },
  {
    "internalType": "uint256",
    "name": "count",
    "type": "uint256"
  },
  {
    "internalType": "uint256",
    "name": "version",
    "type": "uint256"
  },
  {
    "components": [
      {
        "internalType": "bool",
        "name": "isDataOnRight",
        "type": "bool"
      },
      {
        "internalType": "uint8",
        "name": "subtreeHeight",
        "type": "uint8"
      },
      {
        "internalType": "uint256",
        "name": "subtreeSize",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "subtreeVersion",
        "type": "uint256"
      },
      {
        "internalType": "bytes32",
        "name": "siblingHash",
        "type": "bytes32"
      }
    ],
    "internalType": "struct IAVLMerklePath.Data[]",
    "name": "merklePaths",
    "type": "tuple[]"
  }
]
`)
