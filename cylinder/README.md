# Cylinder

The cylinder program is designed for selected members involved in threshold signature scheme (TSS) message signing.

# Introduction

The cylinder program is designed for selected members involved in threshold signature scheme (TSS) message signing. It streamlines the process for participants required to sign a TSS message, ensuring secure and efficient collaboration within the signing group.

Key features of cylinder include:

- Nonce Submission: Allows users to submit nonces that are used during the signing process, ensuring proper coordination for message signing.
- Message Signing: Enables users to sign newly requested messages as part of the TSS protocol.

This tool is essential for members who need to maintain constant engagement and coordination during TSS message signing operations.

# Run cylinder on BandChain mainnet

## 1. Installation

The document is written based on the assumption that the program runs on Ubuntu 22.04 LTS.

Before beginning instructions, the following variables should be set to be used in further instructions. Please make sure that these variables are set every time when using the new shell session.

```sh
# Chain ID of the target chain, e.g. bandchain
export CHAIN_ID=<TARGET_CHAIN_ID>
# Wallet name to be used as a granter account, please change this into your name (no whitespace).
export WALLET_NAME=<YOUR_WALLET_NAME>
# The path for the data and configurations of the program are stored in, e.g. $HOME/.cylinder-account1
export CYLINDER_HOME_PATH=<YOUR_TARGET_PATH>
# url for connecting to the target chain, e.g. tcp://localhost:26657
export RPC_URL=<YOUR_NODE_RPC_URL>
```

### Step 1.1: Install prerequisite programs

To install and run cylinder, the following tools and packages are required:

- make, gcc, g++ (can be obtained from the build-essential package on linux)
- wget, curl, openssl for downloading files
- go version 1.22.3

To install required tools, run the following code

```sh
# install required tools
sudo apt-get update && \
sudo apt-get upgrade -y && \
sudo apt-get install -y build-essential curl wget openssl
```

To install Go version 1.22.3, run the following commands

```sh
# Install Go 1.22.3
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
tar xf go1.22.3.linux-amd64.tar.gz
sudo mv go /usr/local/go

# Set Go path to $PATH variable
echo "export PATH=$PATH:/usr/local/go/bin:~/go/bin" >> $HOME/.profile
source ~/.profile
```

run `go version` to check if go is successfully installed. It should display the go version that is installed.

### Step 1.2: Install BandChain and cylinder program

The cylinder program can be installed via cloning the [Github repository](https://github.com/bandprotocol/chain). To install the BandChain executable program and cylinder, run the following commands.

```sh
cd ~

# Clone BandChain version v3.x.x; TODO: fix BandChain version
git clone https://github.com/bandprotocol/chain
cd chain
git fetch && git checkout v3.x.x

# Install binaries to $GOPATH/bin
make install
```

To check if bandd and cylinder programs are successfully installed, run `bandd version` and `cylinder version`

## 2. Post-Installation Configuration

After successfully installing both the bandd and cylinder programs, there are additional configuration steps that need to be completed before running the cylinder program. These steps ensure that the system is properly set up and ready for operation.

### Step 2.1: Provide granter account to the system

Create a new account using the command below.

```sh
bandd keys add $WALLET_NAME --keyring-backend test
```

If you choose to use an existing account, add the `--recover` flag to the command mentioned above.

whether a granter account is created or recovered, please ensure that the account has sufficient tokens to cover transaction fees.

To verify that the new account has been added to the system, run `bandd keys show $WALLET_NAME -a --keyring-backend test` to display the wallet address associated with the account.

### Step 2.2: Configure general settings

Run the following command to set the configuration of the cylinder program and add signer account to the program.

```sh
cylinder config chain-id $CHAIN_ID --home $CYLINDER_HOME_PATH
cylinder config node $RPC_URL --home $CYLINDER_HOME_PATH
cylinder config granter $(bandd keys show $WALLET_NAME -a --keyring-backend test) --home $CYLINDER_HOME_PATH
cylinder config gas-prices "0.0025uband" --home $CYLINDER_HOME_PATH
cylinder config max-messages 10 --home $CYLINDER_HOME_PATH
cylinder config broadcast-timeout "5m" --home $CYLINDER_HOME_PATH
cylinder config rpc-poll-interval "1s" --home $CYLINDER_HOME_PATH
cylinder config max-try 5 --home $CYLINDER_HOME_PATH
cylinder config min-de 20 --home $CYLINDER_HOME_PATH
cylinder config gas-adjust-start 1.6 --home $CYLINDER_HOME_PATH
cylinder config gas-adjust-step 0.2 --home $CYLINDER_HOME_PATH
cylinder config random-secret "$(openssl rand -hex 32)" --home $CYLINDER_HOME_PATH
cylinder config checking-de-interval "5m" --home $CYLINDER_HOME_PATH

cylinder keys add signer1 --home $CYLINDER_HOME_PATH
cylinder keys add signer2 --home $CYLINDER_HOME_PATH
```

below is the meaning of the configuration of the system

```go
type Config struct {
	ChainID          string        // ChainID of the target chain
	NodeURI          string        // Remote RPC URI of BandChain node to connect to
	Granter          string        // The granter address
	GasPrices        string        // Gas prices of the transaction
	LogLevel         string        // Log level of the logger
	MaxMessages      uint64        // The maximum number of messages in a transaction
	BroadcastTimeout time.Duration // The time that cylinder will wait for tx commit
	RPCPollInterval  time.Duration // The duration of rpc poll interval
	MaxTry           uint64        // The maximum number of tries to submit a report transaction
	MinDE            uint64        // The minimum number of DE
	GasAdjustStart   float64       // The start value of gas adjustment
	GasAdjustStep    float64       // The increment step of gad adjustment
	RandomSecret     tss.Scalar    // The secret value that is used for random D,E
}
```

To check that if the signer account is added into the program, run the following command
`cylinder keys list --home $CYLINDER_HOME_PATH`. The configuration is updated in the `$CYLINDER_HOME_PATH/config.yaml`

### Step 2.3: Set grantee and send tokens to the signer account.

Run the following commands to send 1 BAND to the predefined signer accounts and designate them as grantees of the granter account.

```sh
bandd tx multi-send 1000000uband $(cylinder keys list -a --home $CYLINDER_HOME_PATH) --gas-prices 0.0025uband --keyring-backend test --chain-id $CHAIN_ID --from $WALLET_NAME -b sync -y --node $RPC_URL

bandd tx tss add-grantees $(cylinder keys list -a --home $CYLINDER_HOME_PATH) --gas-prices 0.0025uband --keyring-backend test --chain-id $CHAIN_ID --gas 350000 --from $WALLET_NAME -b sync -y --node $RPC_URL
```

## Run the cylinder program

Run the cylinder program using the command line below

```sh
cylinder run --home $CYLINDER_HOME_PATH
```

# Run cylinder on BandChain local network

1. Go to chain directory
2. Run `make install`
3. Run `chmod +x scripts/start_cylinder.sh` to change the access permission of `start_cylinder.sh`
4. Run `./scripts/start_cylinder.sh` to start Cylinder
