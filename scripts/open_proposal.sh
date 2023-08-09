export WALLET_NAME=validator
export CHAIN_ID="bandchain"
export DAEMON_NAME=bandd
export DAEMON_HOME=$HOME/.band

bandd config broadcast-mode block
bandd config chain-id $CHAIN_ID

export UPGRADE_NAME=v2_6
export UPGRADE_HEIGHT=40

bandd tx gov submit-proposal software-upgrade $UPGRADE_NAME \
  --title upgrade \
  --description upgrade \
  --upgrade-height $UPGRADE_HEIGHT \
  --from $WALLET_NAME \
  -y --keyring-backend test --gas-prices 0.0025uband

