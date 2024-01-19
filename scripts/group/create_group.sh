
#!/bin/bash

export WALLET_NAME=group1

BASEDIR=$(dirname "$0")

TX_HASH=$(bandd tx group create-group $(bandd keys show $WALLET_NAME --address --keyring-backend test) "ipfs://" $BASEDIR/group_members.json \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync --output json | jq '.txhash'| tr -d '"')
echo "TX_HASH: $TX_HASH"

sleep 3

GROUP_ID=$(bandd query tx $TX_HASH --output json | jq '.events' | jq -r '.[] | select(.type == "cosmos.group.v1.EventCreateGroup") | .attributes[0].value' | jq -r '.')
echo "GROUP_ID: $GROUP_ID"

