# Schnorr

This package is the modified version from https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr to support the use cases of Bandchain's TSS module.

## Modification
- Adjust r and s in Signature to be public fields
- Adjust r to be Jacobian points to keep both x,y and we won't enforce even y in our TSS.
