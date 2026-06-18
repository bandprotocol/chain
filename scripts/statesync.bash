#!/bin/bash
# microtick and bitcanna contributed significantly here.
set -uxe

# set environment variables
export GOPATH=~/go
export PATH=$PATH:~/go/bin
export RPC=http://rpc.laozi1.bandchain.org:80
export RPCN=http://rpc.laozi1.bandchain.org:80
export APPNAME=BANDD

# Install Band
go install ./...

# MAKE HOME FOLDER AND GET GENESIS
bandd init notional-band-relays
wget -O ~/.sifnoded/config/genesis.json https://raw.githubusercontent.com/bandprotocol/launch/master/laozi-mainnet/genesis.json


INTERVAL=1000

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-INTERVAL))
TRUST_HASH=$(curl -s "$RPC/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export $(echo $APPNAME)_P2P_LADDR=tcp://0.0.0.0:2210
export $(echo $APPNAME)_RPC_LADDR=tcp://0.0.0.0:2211
export $(echo $APPNAME)_GRPC_ADDRESS=0.0.0.0:2212
export $(echo $APPNAME)_GRPC_WEB_ADDRESS=0.0.0.0:2214
export $(echo $APPNAME)_API_ADDRESS=tcp://127.0.0.1:2213
export $(echo $APPNAME)_NODE=tcp://127.0.0.1:2211
export $(echo $APPNAME)_STATESYNC_ENABLE=true
export $(echo $APPNAME)_P2P_MAX_NUM_OUTBOUND_PEERS=500
export $(echo $APPNAME)_STATESYNC_RPC_SERVERS="$RPC,$RPCN"
export $(echo $APPNAME)_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export $(echo $APPNAME)_STATESYNC_TRUST_HASH=$TRUST_HASH
export $(echo $APPNAME)_P2P_SEEDS="8d42bdcb6cced03e0b67fa3957e4e9c8fd89015a@34.87.86.195:26656","543e0cab9c3016a0e99775443a17bcf163038912@34.150.156.78:26656"
           

bandd start 
