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

## Installation

### Binaries

You can find the latest binaries on our [releases](https://github.com/bandprotocol/chain/releases) page.

### Building from source

To install BandChain's daemon `bandd`, you need to have [Go](https://golang.org/) (version 1.13.5 or later) and [gcc](https://gcc.gnu.org/) installed on our machine. Navigate to the Golang project [download page](https://golang.org/dl/) and gcc [install page](https://gcc.gnu.org/install/), respectively for install and setup instructions.

## Running a Validator Node on the Bandchain Mainnet

The following steps shows how to set up a validator node on the Bandchain mainnet. For similar instructions on running a validator node on our testnet, please refer to [this article](https://medium.com/bandprotocol/bandchain-guanyu-testnet-3-successful-upgrade-how-to-join-as-a-validator-2766ca6717d4)

We recommend the following for running a BandChain Validator:

- **2 or more** CPU cores
- **8 GB **of RAM
- At least **256GB** of disk storage

## Setting Up Validator Node

### Downloading the binaries

We will be assuming that you will be running your node on a Ubuntu 18.04 LTS machine that is allowing connections to port 26656.

To start, you’ll need to install the various utility tools and Golang on the machine.

```bash
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install -y build-essential curl wget

wget https://golang.org/dl/go1.14.9.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.14.9.linux-amd64.tar.gz
rm go1.14.9.linux-amd64.tar.gz

echo "export PATH=\$PATH:/usr/local/go/bin:~/go/bin" >> $HOME/.profile
source ~/.profile
```

### Build BandChain Daemon

Next, you will need to clone and build BandChain. The canonical version for this GuanYu Mainnet is v1.2.6.

```bash
git clone https://github.com/bandprotocol/bandchain
cd bandchain/chain
git checkout v1.2.6
make install

# Check that the correction version of bandd is installed
bandd version --long
# Both should show:
# name: bandchain
# server_name: bandd
# version: chain/v1.2.6
# commit: 2689a3ae4b0b866e198ba31fd88c4c300090a49b
# build_tags: ledger
# go: go version go1.14.9 linux/amd64
```

### Creating BandChain Account and Setup Config

Once installed, you can use the `bandd` CLI to create a new BandChain wallet address and initialize the chain. Please make sure to keep your mnemonic safe!

```bash
# Create a new Band wallet. Do not lose your mnemonic!
bandd keys add <YOUR_WALLET>

# Initialize a blockchain environment for generating genesis transaction.
bandd init --chain-id band-guanyu-mainnet <YOUR_MONIKER>
```

You can then download the official genesis file from the repository. You should also add the initial peer nodes to your Tendermint configuration file.

```bash
# Download genesis file from the repository.
wget https://raw.githubusercontent.com/bandprotocol/launch/master/band-guanyu-mainnet/genesis.json
# Check genesis hash
sudo apt-get install jq
jq -S -c -M '' genesis.json | shasum -a 256
# It should get this hash
9673376e8416d1e7fc87d271b8a9e5e3d3ce78a076b0d907d87c782bb0320e30  -
# Move the genesis file to the proper location
mv genesis.json $HOME/.band/config
# Add some persistent peers
sed -E -i \
  's/persistent_peers = \".*\"/persistent_peers = \"924a8094846222e14c7b86bfb42c0ddfd93cc6d1@gyms1.bandchain.org:26656,4c0b2cadc5ec7de90379b4a8fb7e19c252c7e565@gyms2.bandchain.org:26656\"/' \
  $HOME/.band/config/config.toml
```

### Starting the Blockchain Daemon

With all configurations ready, you can start your blockchain node with a single command. In this tutorial, however, we will show you a simple way to set up `systemd` to run the node daemon with auto-restart.

- Create a config file, using the contents below, at `/etc/systemd/system/bandd.service`. You will need to edit the default ubuntu username to reflect your machine’s username. Note that you may need to use sudo as it lives in a protected folder

```
[Unit]
Description=BandChain Node Daemon
After=network-online.target
[Service]
User=ubuntu
ExecStart=/home/ubuntu/go/bin/bandd start
Restart=always
RestartSec=3
LimitNOFILE=4096
[Install]
WantedBy=multi-user.target
```

- Install the service and start the node

```
sudo systemctl enable bandd
sudo systemctl start bandd
```

While not required, it is recommended that you run your validator node behind your sentry nodes for DDOS mitigation. See [this thread](https://forum.cosmos.network/t/sentry-node-architecture-overview/454) for some example setups. Your node will now start connecting to other nodes and syncing the blockchain state.

### ⚠️ Wait Until Your Chain is Fully Sync

You can tail the log output with `journalctl -u bandd.service -f`. If all goes well, you should see that the node daemon has started syncing. Now you should wait until your node has caught up with the most recent block.

```shell
... bandd: I[..] Executed block  ... module=state height=20000 ...
... bandd: I[..] Committed state ... module=state height=20000 ...
... bandd: I[..] Executed block  ... module=state height=20001 ...
... bandd: I[..] Committed state ... module=state height=20001 ...
```

See the our [explorer](https://cosmoscan.io/) for the height of the latest block. Syncing should take a while, depending on your internet connection.

⚠️ **NOTE:** You should not proceed to the next step until your node caught up to the latest block.

### Send Yourself BAND Token

With everything ready, you will need some BAND tokens to apply as a validator. You can use `bandd` keys list command to show your address.

```shell
bandd keys list
- name: ...
  type: local
  address: band1g3fd6rslryv498tjqmmjcnq5dlr0r6udm2rxjk
  pubkey: ...
  mnemonic: ""
  threshold: 0
  pubkeys: []
```

### Apply to Become Block Validator

Once you have some BAND tokens, you can apply to become a validator by sending `MsgCreateValidator` transaction.

```bash
bandd tx staking create-validator \
    --amount <your-amount-to-stake>uband \
    --commission-max-change-rate 0.01 \
    --commission-max-rate 0.2 \
    --commission-rate 0.1 \
    --from <your-wallet-name> \
    --min-self-delegation 1 \
    --moniker <your-moniker> \
    --pubkey $(bandd tendermint show-validator) \
    --chain-id band-guanyu-mainnet
```

Once the transaction is mined, you should see yourself on the [validator page](https://cosmoscan.io/validators). Congratulations. You are now a working BandChain mainnet validator!

### Setting Up Yoda — The Oracle Daemon

For Phase 1, BandChain validators are also responsible for responding to oracle data requests. Whenever someone submits a request message to BandChain, the chain will autonomously choose a subset of active oracle validators to perform the data query.

The validators are chosen submit a report message to BandChain within a given timeframe as specified by a chain parameter (100 blocks in Guanyu mainnet). We provide a program called yoda to do this task for you. For more information on the data request process, please see [here](https://docs.bandchain.org/whitepaper/system-overview.html#oracle-data-request).

Yoda uses an external executor to resolve requests to data sources. Currently, it supports [AWS Lambda](https://aws.amazon.com/lambda/) (through the REST interface).

In future releases, `yoda` will support more executors and allow you to specify multiple executors to add redundancy. Please use [this link](https://github.com/bandprotocol/bandchain/wiki/AWS-lambda-executor-setup) to setup lambda function.

You also need to set up `yoda` and activate oracle status. Here’s the [documentation](https://github.com/bandprotocol/bandchain/wiki/Instruction-for-apply-to-be-an-oracle-validator-on-Guanyu-mainnet) to get started.

That’s it! You can verify that your validator is now an oracle provider on the [block explorer](https://cosmoscan.io). Your yoda process must be responding to oracle requests assigned to your node. If the process misses a request, your oracle provider status will automatically get deactivated and you must send MsgActivate to activate again after a 10-minute waiting period and make sure that yoda is up.

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
    - [CosmoScan Testnet](https://guanyu-testnet3.cosmoscan.io)

## Community

- [Official Website](https://bandprotocol.com)
- [Telegram](https://100.band/tg)
- [Twitter](https://twitter.com/bandprotocol)
- [Developer Discord](https://100x.band/discord)

## License & Contributing

BandChain is licensed under the terms of the GPL 3.0 License unless otherwise specified in the LICENSE file at module's root.

We highly encourage participation from the community to help with D3N development. If you are interested in developing with D3N or have suggestion for protocol improvements, please open an issue, submit a pull request, or [drop as a line].

[drop as a line]: mailto:connect@bandprotocol.com
