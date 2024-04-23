#!/bin/bash

export WALLET_NAME=requester
export TEXT=62616e6470726f746f636f6c
export GROUP_ID=1

bandd tx bandtss request-signature text $TEXT \
    --from requester --keyring-backend test \
    --gas-prices 0.0025uband --fee-limit 100uband \
    -b sync -y
