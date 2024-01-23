
#!/bin/bash

export ADDRESS1=group1
export ADDRESS2=group2

BASEDIR=$(dirname "$0")

TX_HASH=$(bandd tx group submit-proposal $BASEDIR/proposal_update_symbols.json \
  --from $ADDRESS1 \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync --output json | jq '.txhash'| tr -d '"')
echo "TX_HASH: $TX_HASH"

sleep 3

PROPOSAL_ID=$(bandd query tx $TX_HASH --output json | jq '.events' | jq -r '.[] | select(.type == "cosmos.group.v1.EventSubmitProposal") | .attributes[0].value' | jq -r '.')
echo "PROPOSAL_ID: $PROPOSAL_ID"

# Vote and exec
bandd tx group vote $PROPOSAL_ID $(bandd keys show $ADDRESS1 --address --keyring-backend test) VOTE_OPTION_YES "agree"  --from $ADDRESS1 -y --keyring-backend test --gas-prices 0.0025uband -b sync
sleep 3
bandd tx group vote $PROPOSAL_ID $(bandd keys show $ADDRESS2 --address --keyring-backend test) VOTE_OPTION_YES "agree"  --from $ADDRESS2 -y --keyring-backend test --gas-prices 0.0025uband -b sync
sleep 3
bandd tx group exec $PROPOSAL_ID --from $ADDRESS1 -y --keyring-backend test --gas-prices 0.0025uband -b sync
