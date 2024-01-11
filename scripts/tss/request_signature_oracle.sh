#!/bin/bash

export WALLET_NAME=requester
export REQUEST_ID=1
export GROUP_ID=1
export ENCODE_TYPE=1

bandd tx tss request-signature oracle-result $REQUEST_ID $ENCODE_TYPE \
    --group-id $GROUP_ID \
    --from requester --keyring-backend test \
    --gas-prices 0.0025uband --fee-limit 100uband \
    -b sync -y
