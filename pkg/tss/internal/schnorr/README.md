# Schnorr

This package is the modified version from https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr to support the use cases of Bandchain's TSS module.

## Modification
- Adjust r and s in Signature to be public fields
- Adjust r to be Jacobian points to keep both x and y since we won't enforce even y in our TSS.
- Add a complaint signature to keep A1, A2, and Z. 
- Adjust ComputeSignatureS and Verify to be compatible with our TSS.
