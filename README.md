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
  <br/>
  <a href="http://docs.bandchain.org/whitepaper/introduction.html">Whitepaper</a>
  ·
  <a href="http://docs.bandchain.org/technical-specifications/obi.html">Technical Specifications</a>
  ·
  <a href="http://docs.bandchain.org/using-any-datasets/">Developer Documentation</a>
  ·
  <a href="http://docs.bandchain.org/client-library/data.html">Client Library</a>
</p>

<br/>

## What is BandChain?

BandChain is a **cross-chain data oracle platform** that aggregates and connects real-world data and APIs to smart contracts. It is designed to be **compatible with most smart contract and blockchain development frameworks**. It does the heavy lifting jobs of pulling data from external sources, aggregating them, and packaging them into the format that’s easy to use and verified efficiently across multiple blockchains.

Band's flexible oracle design allows developers to **query any data** including real-world events, sports, weather, random numbers and more. Developers can create custom-made oracles using WebAssembly to connect smart contracts with traditional web APIs within minutes.

## Installation

### Building from source

We recommend the following for running a BandChain Validator:

- **2 or more** CPU cores
- **8 GB** of RAM (16 GB in case on participate in mainnet upgrade)
- At least **100GB** of disk storage

**Step 1. Install Golang**

Go v1.15+ or higher is required for BandChain.

If you haven't already, install Golang by following the [official docs](https://golang.org/doc/install). Make sure that your GOPATH and GOBIN environment variables are properly set up.

**Step 2. Get BandChain source code**

Use `git` to retrieve BandChain from the [official repo](https://github.com/bandprotocol/chain), and checkout the master branch, which contains the latest stable release. That should install the `bandd` binary.

```bash
git clone https://github.com/bandprotocol/chain
git checkout master
make install
```

**Step 3. Verify your installation**

Using `bandd version` command to verify that your `bandd` has been build successfully.

```
bandd version --long
name: bandchain
server_name: bandd
version: [Latest version]
commit: 6f5d4742d311d1bcc57d93f3c37812dc59292212
build_tags: ledger
go: go version go1.15.7 darwin/amd64
build_deps:
...
```

### Setting Up Yoda — The Oracle Daemon

BandChain validators are also responsible for responding to oracle data requests. Whenever someone submits a request message to BandChain, the chain will autonomously choose a subset of active oracle validators to perform the data query.

The validators are chosen submit a report message to BandChain within a given timeframe as specified by a chain parameter (100 blocks in mainnet). We provide a program called yoda to do this task for you. For more information on the data request process, please see [here](https://docs.bandchain.org/whitepaper/system-overview.html#oracle-data-request).

Yoda uses an external executor to resolve requests to data sources. Currently, it supports [AWS Lambda](https://aws.amazon.com/lambda/) and [Google Cloud Function](https://cloud.google.com/functions) (through the REST interface). In future releases, `yoda` will support more executors and allow you to specify multiple executors to add redundancy.

You also need to set up `yoda` and activate oracle status. Here’s the [documentation](https://github.com/bandprotocol/bandchain/wiki/Instruction-for-apply-to-be-an-oracle-validator-on-Guanyu-mainnet) to get started.

That’s it! You can verify that your validator is now an oracle provider via cli by using ` bandd query oracle validator <your validator address>`. Your yoda process must be responding to oracle requests assigned to your node. If the process misses a request, your oracle provider status will automatically get deactivated and you must send MsgActivate to activate again after a 10-minute waiting period and make sure that yoda is up.

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
    - [CosmoScan Testnet](https://laozi-testnet2.cosmoscan.io)

## Community

- [Official Website](https://bandprotocol.com)
- [Telegram](https://100.band/tg)
- [Twitter](https://twitter.com/bandprotocol)
- [Developer Discord](https://100x.band/discord)

## License & Contributing

BandChain is licensed under the terms of the GPL 3.0 License unless otherwise specified in the LICENSE file at module's root.

We highly encourage participation from the community to help with D3N development. If you are interested in developing with D3N or have suggestion for protocol improvements, please open an issue, submit a pull request, or [drop as a line].

[drop as a line]: mailto:connect@bandprotocol.com
