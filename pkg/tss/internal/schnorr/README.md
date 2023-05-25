# Schnorr

This package is the modified version from https://github.com/decred/dcrd/tree/master/dcrec/secp256k1/schnorr to support the use cases of Bandchain's TSS module.

## Modification
- Change R from FieldVal to Jacobian Point by using compress/uncompress from PublicKey to support x and y.
- Remove hashing function for Schnorr functions and requires the caller function to hash by itself.
- Change function from `schnorrSign` to `ComputeSigS` and only calculate S of signature.
- Change function from `schnorrVerify` to `Verify` with the custom generator parameter (Default is G).
- Modify `Verify` function to receive R and S from the signature separately (instead of r in the signature).
- Update test cases to support the modified version.
