#!/bin/bash

HOME_PATH="$HOME/.cylinder"
if ! [ -z "$1" ]; then
  HOME_PATH=$HOME_PATH-$1
fi

KEY="${1:-validator}"

echo "HomePath: $HOME_PATH"
echo "Key: $KEY"

rm -rf $HOME_PATH

# config chain id
cylinder config chain-id bandchain --home $HOME_PATH

# add member to cylinder config
cylinder config granter $(bandd keys show $KEY -a --keyring-backend test) --home $HOME_PATH

# setup broadcast-timeout to cylinder config
cylinder config broadcast-timeout "5m" --home $HOME_PATH

# setup rpc-poll-interval to cylinder config
cylinder config rpc-poll-interval "1s" --home $HOME_PATH

# setup max-try to cylinder config
cylinder config max-try 5 --home $HOME_PATH

# setup gas-prices to cylinder config
cylinder config gas-prices "0.0025uband" --home $HOME_PATH

for i in $(eval echo {1..2})
do
  # add signer key
  cylinder keys add signer$i --home $HOME_PATH
done

# send band tokens to grantees
bandd tx multi-send 1000000uband $(cylinder keys list -a --home $HOME_PATH) --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain --from $KEY -b block -y

# wait for sending band tokens transaction success
sleep 2

bandd tx tss add-grantees $(cylinder keys list -a --home $HOME_PATH) --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain --from $KEY -b block -y 

# wait for adding gratees transaction success
sleep 2

# run cylinder
cylinder run --home $HOME_PATH
