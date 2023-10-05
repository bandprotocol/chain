CURR_HEIGHT=$(bandd query block | jq -r '.block.header.height')

UPGRADE_HEIGHT=$(($CURR_HEIGHT+60))

TX_HASH=$(bandd tx gov submit-proposal software-upgrade $1 \
  --title upgrade --upgrade-info upgrade \
  --description upgrade \
  --upgrade-height $UPGRADE_HEIGHT --deposit 1000000000uband \
  --from requester --chain-id bandchain \
  -y --keyring-backend test --gas-prices 0.0025uband -b sync --output json| jq '.txhash'| tr -d '"')

sleep 3

PROPOSAL_ID=$(bandd query tx $TX_HASH --output=json | jq -r '.raw_log | fromjson | .[] | .events[] | select(.type=="submit_proposal") | .attributes[] | select(.key=="proposal_id") | .value')

bandd tx gov vote $PROPOSAL_ID VOTE_OPTION_YES --from validator --keyring-backend test --gas-prices 0.0025uband -y -b sync --chain-id bandchain
