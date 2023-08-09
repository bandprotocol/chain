export PROPOSAL_ID=1
export WALLET_NAME=validator
export CHAIN_ID="bandchain"

bandd config broadcast-mode block
bandd config chain-id $CHAIN_ID

bandd tx gov deposit $PROPOSAL_ID 1000000000uband \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband

bandd tx gov vote $PROPOSAL_ID yes \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband
