#!/bin/bash

export WALLET_NAME=requester
export REQUEST_ID=1
export ENCODER=1

bandd tx bandtss request-signature oracle-result $REQUEST_ID $ENCODER \
    --from requester --keyring-backend test \
    --gas-prices 0.0025uband --fee-limit 100uband \
    -b sync -y
