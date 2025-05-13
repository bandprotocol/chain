CURR_HEIGHT=$(bandd query block --output json | tail -n +2 | jq -r '.header.height')

UPGRADE_HEIGHT=$(($CURR_HEIGHT+60))

jq --arg name "$1" '.messages[0].plan.name = $name' ./scripts/test_script/3rc2_proposal.json > tmp.json && mv tmp.json ./scripts/test_script/3rc2_proposal.json
jq --arg height "$UPGRADE_HEIGHT" '.messages[0].plan.height = $height' ./scripts/test_script/3rc2_proposal.json > tmp.json && mv tmp.json ./scripts/test_script/3rc2_proposal.json

TX_HASH=$(bandd tx gov submit-proposal ./scripts/test_script/3rc2_proposal.json \
  --from requester --chain-id bandchain \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync --output json| jq '.txhash'| tr -d '"')

sleep 3

PROPOSAL_ID=$(bandd query tx $TX_HASH --output=json | tail | jq -r '.events[] | select(.type=="submit_proposal") | .attributes[] | select(.key=="proposal_id") | .value')

bandd tx gov vote $PROPOSAL_ID yes --from validator --keyring-backend test --gas-prices 0.0025uband -y -b sync --chain-id bandchain


