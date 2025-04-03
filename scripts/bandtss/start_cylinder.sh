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

# setup max-messages to cylinder config
cylinder config max-messages 20 --home $HOME_PATH

# setup broadcast-timeout to cylinder config
cylinder config broadcast-timeout "5m" --home $HOME_PATH

# setup rpc-poll-interval to cylinder config
cylinder config rpc-poll-interval "1s" --home $HOME_PATH

# setup max-try to cylinder config
cylinder config max-try 5 --home $HOME_PATH

# setup gas-prices to cylinder config
cylinder config gas-prices "0uband" --home $HOME_PATH

# setup min-de to cylinder config
cylinder config min-de 100 --home $HOME_PATH

# setup gas-adjust-start to cylinder config
cylinder config gas-adjust-start 1.6 --home $HOME_PATH

# setup gas-adjust-step to cylinder config
cylinder config gas-adjust-step 0.2 --home $HOME_PATH

# setup random-secret to cylinder config
cylinder config random-secret "$(openssl rand -hex 32)" --home $HOME_PATH

# setup checking DE interval to cylinder config
cylinder config checking-de-interval "1m" --home $HOME_PATH

for i in $(eval echo {1..4})
do
  # add signer key
  cylinder keys add signer$i --home $HOME_PATH
done

# send band tokens to grantees
bandd tx bank multi-send $KEY $(cylinder keys list -a --home $HOME_PATH) 1000000uband --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain --from $KEY -b sync -y

# wait for sending band tokens transaction success
sleep 6

bandd tx tss add-grantees $(cylinder keys list -a --home $HOME_PATH) --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain --gas 700000 --from $KEY -b sync -y 

sleep 6

# run cylinder
cylinder run --home $HOME_PATH --metrics-listen-addr ":8082"
