
#!/bin/bash

export WALLET_NAME=validator1
export ENDPOINT=http://localhost:26657

BASEDIR=$(dirname "$0")

# Submit create_group proposal
bandd tx gov submit-proposal $BASEDIR/proposal_update_params.json \
  --from $WALLET_NAME \
  --node $ENDPOINT \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync

sleep 5

# Query last proposal's id
PROPOSAL_ID=$(bandd query gov proposals --node $ENDPOINT --reverse --limit 1 --output json | jq -r '.proposals[0].id')

# Vote on that proposal
echo "...Voting to proposal $PROPOSAL_ID..."
bandd tx gov vote $PROPOSAL_ID yes \
  --from $WALLET_NAME \
  --node $ENDPOINT \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync
