
#!/bin/bash

export WALLET_NAME=group1
export GROUP_ID=1

BASEDIR=$(dirname "$0")

TX_HASH=$(bandd tx group create-group-policy $(bandd keys show $WALLET_NAME --address --keyring-backend test) $GROUP_ID "{\"name\":\"policy 1\",\"description\":\"\"}" $BASEDIR/threshold_policy.json \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync --output json | jq '.txhash'| tr -d '"')
echo "TX_HASH: $TX_HASH"

sleep 3

GROUP_POLICY_ADDRESS=$(bandd query tx $TX_HASH --output json | jq '.events' | jq -r '.[] | select(.type == "cosmos.group.v1.EventCreateGroupPolicy") | .attributes[0].value' | jq -r '.')
echo "GROUP_POLICY_ADDRESS: $GROUP_POLICY_ADDRESS"
