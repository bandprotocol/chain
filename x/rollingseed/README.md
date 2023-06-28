# `rollingseed`

## Abstract

This abstract outlines a module designed to enhance the security and randomness of Cosmos-SDK based blockchains by implementing an on-chain rolling seed system. The module introduces a mechanism whereby a new rolling seed is generated at the beginning of each block, utilizing the block hash as a source of entropy. By integrating this module into a Cosmos-SDK based blockchain, developers can ensure a higher level of security and prevent predictability in critical processes such as block validation, leader selection, cryptographic operations, and on-chain random number generation (RNG).

One of the significant use cases for this module is the on-chain RNG functionality it enables. Random numbers play a crucial role in various applications such as gaming, gambling, and fair distributed systems. By leveraging the on-chain rolling seed system, developers can create a secure and transparent on-chain RNG. The generation of a new rolling seed based on the block hash ensures that the RNG output is unpredictable and tamper-proof. This capability opens up opportunities for provably fair games, random selection of validators or winners, and other use cases where unbiased randomness is essential.

This module will be used in the Oracle and TSS modules in BandChain.
