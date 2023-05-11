#!/bin/bash

rm -rf ~/.cylinder

# config chain id
cylinder config chain-id bandchain

# add member to cylinder config
cylinder config granter $(bandd keys show validator -a --keyring-backend test)

# setup broadcast-timeout to cylinder config
cylinder config broadcast-timeout "5m"

# setup rpc-poll-interval to cylinder config
cylinder config rpc-poll-interval "1s"

# setup max-try to cylinder config
cylinder config max-try 5

# setup gas-prices to cylinder config
cylinder config gas-prices "0.0025uband"

# wait for activation transaction success
sleep 2

for i in $(eval echo {1..2})
do
  # add signer key
  cylinder keys add signer$i
done

# send band tokens to grantees
bandd tx multi-send 1000000uband $(cylinder keys list -a) --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain --from validator -y

# wait for sending band tokens transaction success
sleep 2

bandd tx tss add-grantees $(cylinder keys list -a) --from validator --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b block -y 

# wait for adding gratees transaction success
sleep 2

# run cylinder
cylinder run
