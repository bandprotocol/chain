#!/bin/bash

export WALLET_NAME=requester
export SIGNAL_IDs=crypto_price.ethusd,crypto_price.usdtusd

# 0: Unspecified, 1: Default, 2: Tick
export FEEDS_TYPE=1

bandd tx bandtss request-signature feeds-prices $SIGNAL_IDs $FEEDS_TYPE \
    --from $WALLET_NAME --keyring-backend test \
    --gas-prices 0.0025uband --fee-limit 100uband \
    -b sync -y
