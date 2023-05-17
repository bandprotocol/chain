# Schnorr

This package is the modified version from https://github.com/decred/dcrd/tree/master/dcrec/secp256k1/schnorr to support the use cases of Bandchain's TSS module.

## Modification
- Don't limit the message to 32 bytes as we will hash it later with the nonce.
- Use Keccak256 instead of blake256 for the hash generation to support the signature on the EVM chain.
- Change function from `schnorrSign` to `Sign`.
- Change function from `schnorrVerify` to `Verify` with the custom generator parameter (Default is G).
- Modify `Verify` function to support override r in case we want to compare the final result with other values (instead of r in the signature).
- Update test cases to support the modified version.
