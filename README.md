<p>&nbsp;</p>
<p align="center">

<img src="./odinprotocol_logo.png" width="500px" alt="odin logo">

</p>

<p align="center">OdinChain - Decentralized Data Delivery Network<br/><br/>

<a href="https://pkg.go.dev/badge/github.com/GeoDB-Limited/odin-core">
    <img src="https://pkg.go.dev/badge/github.com/GeoDB-Limited/odin-core">
</a>
<a href="https://goreportcard.com/badge/github.com/GeoDB-Limited/odin-core">
    <img src="https://goreportcard.com/badge/github.com/GeoDB-Limited/odin-core">
</a>
<a href="https://github.com/GeoDB-Limited/odin-core/workflows/Tests/badge.svg">
    <img src="https://github.com/GeoDB-Limited/odin-core/workflows/Tests/badge.svg">
</a>

<p align="center">
  <a href="https://app.gitbook.com/@geodb/s/odin-protocol/"><strong>Documentation »</strong></a>
  <br />
  <br/>
  <a href="https://odinprotocol.io/docs/odin-whitepaper.pdf">Whitepaper</a> | 
  <a href="https://odinprotocol.io/docs/odin-tokenomics.pdf">Tokenomics paper</a>
</p>

<br/>

_Current TestNet name is "**vidar** - another son of the supreme god Odin and Grid (a giantess), and his powers were
matched only by that of Thor."_ <br>
_Name:_ **odin-testnet-vidar**

## Installation

### Binaries

You can find the latest binaries on our [releases](https://github.com/GeoDB-Limited/odin-core/releases) page.

### Building from source

To install OdinChain's daemon `bandd`, you need to have [Go](https://golang.org/) (version 1.13.5 or later)
and [gcc](https://gcc.gnu.org/) installed on our machine. Navigate to the Golang
project [download page](https://golang.org/dl/) and gcc [install page](https://gcc.gnu.org/install/), respectively for
install and setup instructions.

## Running a Validator Node on the OdinChain TestNet

The following steps shows how to set up a validator node on the odinchain testnet. For similar instructions on running a
validator node on our testnet, please refer
to [this article](https://medium.com/odinprotocol/odinchain-guanyu-testnet-3-successful-upgrade-how-to-join-as-a-validator-2766ca6717d4)

We recommend the following for running a odinChain Validator:

- **2 or more** CPU cores
- **8 GB **of RAM
- At least **256GB** of disk storage

## Setting Up Validator Node

### Downloading the binaries

We will be assuming that you will be running your node on a Ubuntu 18.04 LTS machine that is allowing connections to
port 26656.

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

### Build OdinChain Daemon

Next, you will need to clone and build OdinChain. The canonical version for this GuanYu Mainnet is v1.2.6.

```bash
git clone https://github.com/GeoDB-Limited/odin-core
cd odin-core
git checkout testnet-{name}
make install

# Check that the correction version of bandd is installed
bandd version --long

### Creating OdinChain Account and Setup Config

Once installed, you can use the `bandd` CLI to create a new OdinChain wallet address and initialize the chain. Please make sure to keep your mnemonic safe!

```bash
# Create a new odin wallet. Do not lose your mnemonic!
bandd keys add <YOUR_WALLET>

# Initialize a blockchain environment for generating genesis transaction.
bandd init --chain-id odin-testnet-{name} <YOUR_MONIKER>
```

You can then download the official genesis file from the repository. You should also add the initial peer nodes to your
Tendermint configuration file.

```bash
# Download genesis file from the repository.
wget https://raw.githubusercontent.com/GeoDB-Limited/odin-core/master/testnets/odin-testnet-{name}/genesis.json
# Check genesis hash
sudo apt-get install jq
# Move the genesis file to the proper location
mv genesis.json $HOME/.odin/config
# Add some persistent peers
sed -E -i \
  's/persistent_peers = \".*\"/persistent_peers = \"11392b605378063b1c505c0ab123f04bd710d7d7@node.testnet.odinprotocol.io/asgard/service/' \
  $HOME/.odin/config/config.toml
```

### Starting the Blockchain Daemon

With all configurations ready, you can start your blockchain node with a single command. In this tutorial, however, we
will show you a simple way to set up `systemd` to run the node daemon with auto-restart.

- Create a config file, using the contents below, at `/etc/systemd/system/bandd.service`. You will need to edit the
  default ubuntu username to reflect your machine’s username. Note that you may need to use sudo as it lives in a
  protected folder

```
[Unit]
Description=odinChain Node Daemon
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

While not required, it is recommended that you run your validator node behind your sentry nodes for DDOS mitigation.
See [this thread](https://forum.cosmos.network/t/sentry-node-architecture-overview/454) for some example setups. Your
node will now start connecting to other nodes and syncing the blockchain state.

### ⚠️ Wait Until Your Chain is Fully Sync

You can tail the log output with `journalctl -u bandd.service -f`. If all goes well, you should see that the node daemon
has started syncing. Now you should wait until your node has caught up with the most recent block.

```bash
... bandd: I[..] Executed block  ... module=state height=20000 ...
... bandd: I[..] Committed state ... module=state height=20000 ...
... bandd: I[..] Executed block  ... module=state height=20001 ...
... bandd: I[..] Committed state ... module=state height=20001 ...
```

⚠️ **NOTE:** You should not proceed to the next step until your node caught up to the latest block.

### Send Yourself odin Token

With everything ready, you will need some odin tokens to apply as a validator. You can use `bandd` keys list command to
show your address.

```bash
bandd keys list
- name: ...
  type: local
  address: odin1g3fd6rslryv498tjqmmjcnq5dlr0r6udm2rxjk
  pubkey: ...
  mnemonic: ""
  threshold: 0
  pubkeys: []
```

### Apply to Become Block Validator

Once you have some odin tokens, you can apply to become a validator by sending `MsgCreateValidator` transaction.

```bash
bandd tx staking create-validator \
    --amount <your-amount-to-stake>odin \
    --commission-max-change-rate 0.01 \
    --commission-max-rate 0.2 \
    --commission-rate 0.1 \
    --from <your-wallet-name> \
    --min-self-delegation 1 \
    --moniker <your-moniker> \
    --pubkey $(odind tendermint show-validator) \
    --chain-id odin-guanyu-mainnet
```

Once the transaction is mined, you should see yourself on the [validator page](https://testnet.odinprotocol.io/validators).
Congratulations. You are now a working OdinChain testnet validator!

### Setting Up Yoda — The Oracle Daemon

For Phase 1, OdinChain validators are also responsible for responding to oracle data requests. Whenever someone submits
a request message to OdinChain, the chain will autonomously choose a subset of active oracle validators to perform the
data query.

The validators are chosen submit a report message to OdinChain within a given timeframe as specified by a chain
parameter. We provide a program called yoda to do this task for you.

Yoda uses an external executor to resolve requests to data sources. Currently, it
supports [AWS Lambda](https://aws.amazon.com/lambda/) (through the REST interface).

In future releases, `yoda` will support more executors and allow you to specify multiple executors to add redundancy.
Please use [this link](https://github.com/odinprotocol/odinchain/wiki/AWS-lambda-executor-setup) to setup lambda
function.

## Resources

- Peers:
    - Testnet:
        - [Odin Testnet](https://node.testnet.odinprotocol.io)

## Community

- [Official Website](https://odinprotocol.io)

## License & Contributing

...
