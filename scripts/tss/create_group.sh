#!/bin/bash

export WALLET_NAME=validator

BASEDIR=$(dirname "$0")

bandd tx gov submit-proposal $BASEDIR/create_group_proposal.json \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync

sleep 5

bandd tx gov vote 1 yes \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync
