#!/bin/bash

rm -rf ~/.yoda3

# config chain id
yoda3 config chain-id bandchain

# add validator to yoda config
yoda3 config validator $(bandd keys show validator -a --bech val --keyring-backend test)

# setup execution endpoint
yoda3 config executor "rest:https://asia-southeast2-band-playground.cloudfunctions.net/test-runtime-executor?timeout=10s"

# setup broadcast-timeout to yoda config
yoda3 config broadcast-timeout "5m"

# setup rpc-poll-interval to yoda config
yoda3 config rpc-poll-interval "1s"

# setup max-try to yoda config
yoda3 config max-try 5

echo "y" | bandd tx oracle activate --from validator --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain

# wait for activation transaction success
sleep 2

for i in $(eval echo {1..1})
do
  # add reporter key
  yoda3 keys add reporter$i
done

# send band tokens to reporters
echo "y" | bandd tx bank send validator $(yoda3 keys list -a) 1000000uband --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain

# wait for sending band tokens transaction success
sleep 2

# add reporter to bandchain
echo "y" | bandd tx oracle add-reporters $(yoda3 keys list -a) --from validator --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain

# wait for adding reporter transaction success
sleep 2

# run yoda
yoda3 run
