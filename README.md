<p>&nbsp;</p>
<p align="center">

<img src="bandprotocol_logo.svg" width=500>

</p>

<p align="center">
BandChain - Decentralized Data Delivery Network<br/><br/>

<a href="https://pkg.go.dev/badge/github.com/bandprotocol/chain">
    <img src="https://pkg.go.dev/badge/github.com/bandprotocol/chain">
</a>
<a href="https://goreportcard.com/badge/github.com/bandprotocol/chain">
    <img src="https://goreportcard.com/badge/github.com/bandprotocol/chain">
</a>
<a href="https://github.com/bandprotocol/chain/workflows/Tests/badge.svg">
    <img src="https://github.com/bandprotocol/chain/workflows/Tests/badge.svg">
</a>

<p align="center">
  <a href="https://docs.bandchain.org/"><strong>Documentation »</strong></a>
  <br />
</p>

<br/>

## What is BandChain?

BandChain is a **cross-chain data oracle platform** that aggregates and connects real-world data and APIs to smart contracts. It is designed to be **compatible with most smart contract and blockchain development frameworks**. It does the heavy lifting jobs of pulling data from external sources, aggregating them, and packaging them into the format that’s easy to use and verifiable efficiently across multiple blockchains.

Band's flexible oracle design allows developers to **query any data** including real-world events, sports, weather, random numbers and more. Developers can create custom-made oracles using WebAssembly to connect smart contracts with traditional web APIs within minutes.

## Installation

Please refer to [this documentation](https://docs.bandchain.org/node-validators/run-node/joining-mainnet/installation) for the most up-to-date installation guide.

## Building from source

We recommend the following for running a BandChain Validator:

- **4 or more** CPU cores
- **16 GB** of RAM (32 GB in case on participate in mainnet)
- At least **150GB** of disk storage

**Step 1. Install Golang**

Go v1.24+ or higher is required for BandChain.

If you haven't already, install Golang by following the [official docs](https://golang.org/doc/install). Make sure that your GOPATH and GOBIN environment variables are properly set up.

**Step 2. Get BandChain source code**

Use `git` to retrieve BandChain from the [official repo](https://github.com/bandprotocol/chain), and checkout the master branch, which contains the latest stable release. That should install the `bandd` binary.

```bash
git clone https://github.com/bandprotocol/chain
cd chain && git checkout v3.1.0
make install
```

**Step 3. Verify your installation**

Using `bandd version` command to verify that your `bandd` has been built successfully.

```
name: bandchain
server_name: bandd
version: 3.1.0
build_tags: netgo,ledger
go: go version go1.24.2 darwin/amd64
build_deps:
...
```

If you are using Mac ARM architecture (M1, M2) and face the issue of GMP library, you can run this.
```
brew update && brew install gmp
sudo ln -s /opt/homebrew/lib/libgmp.10.dylib /usr/local/lib/
```

## Useful scripts for development

- `scripts/generate_genesis.sh` to create/reset the default genesis file 
- `scripts/start_bandd.sh` to start the bandd binary
- `scripts/start_yoda.sh` to start yoda with reporter(s)
- `scripts/start_grogu.sh` to start grogu with feeder(s)
- `scripts/bandtss/start_cylinder.sh` to start cylinder with grantee(s)

## Resources

- Developer
  - Documentation: [docs.bandchain.org](https://docs.bandchain.org)
  - SDKs:
    - JavaScript: [bandchainjs](https://www.npmjs.com/package/@bandprotocol/bandchain.js)
    - Python: [pyband](https://pypi.org/project/pyband/)
- Block Explorers:
  - Mainnet:
    - [Cosmoscan Mainnet](https://cosmoscan.io)
    - [Big Dipper](https://band.bigdipper.live/)
  - Testnet:
    - [CosmoScan Testnet](https://band-v3-testnet.cosmoscan.io)

## Community

- [Official Website](https://bandprotocol.com)
- [Telegram](https://t.me/bandprotocol)
- [Twitter](https://twitter.com/bandprotocol)
- [Developer Discord](https://discord.com/invite/3t4bsY7)

## License & Contributing

BandChain is licensed under the terms of the GPL 3.0 License unless otherwise specified in the LICENSE file at module's root.

We highly encourage participation from the community to help with D3N development. If you are interested in developing with D3N or have suggestions for protocol improvements, please open an issue, submit a pull request, or [drop us a line].

[drop us a line]: mailto:connect@bandprotocol.com
