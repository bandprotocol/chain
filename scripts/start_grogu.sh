#!/bin/bash

rm -rf ~/.grogu

# config chain id
grogu config chain-id bandchain

# add validator to grogu config
grogu config validator $(bandd keys show validator -a --bech val --keyring-backend test)

# setup execution endpoint
grogu config bothan "localhost:50051"

# setup broadcast-timeout to grogu config
grogu config broadcast-timeout "5m"

# setup rpc-poll-interval to grogu config
grogu config rpc-poll-interval "1s"

# setup max-try to grogu config
grogu config max-try 5

grogu config log-level debug

grogu config nodes "tcp://localhost:26657,tcp://localhost:26657,tcp://localhost:26657"

echo "y" | bandd tx oracle activate --from validator --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain

# wait for activation transaction success
sleep 2

for i in $(eval echo {1..1})
do
  # add reporter key
  grogu keys add reporter$i
done

# send band tokens to reporters
echo "y" | bandd tx bank send validator $(grogu keys list -a) 1000000uband --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain

# wait for sending band tokens transaction success
sleep 2

# add reporter to bandchain
echo "y" | bandd tx feeds add-grantees $(grogu keys list -a) --from validator --gas-prices 0.0025uband --keyring-backend test --chain-id bandchain

# wait for addding reporter transaction success
sleep 2

# run grogu
grogu run
