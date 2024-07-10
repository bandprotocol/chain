
# delegate
bandd tx staking delegate bandvaloper1p40yh3zkmhcv0ecqp3mcazy83sa57rgjde6wec 100uband --from requester --keyring-backend test --gas-prices 0.0025uband -y

# lock
bandd tx restake lock-power test 80 --from validator --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4
bandd tx restake lock-power test 80 --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4
bandd tx restake lock-power test2 90 --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4

# add-rewards
bandd tx restake add-rewards test 100uband --from requester --keyring-backend test --gas-prices 0.0025uband -y
sleep 4
bandd tx restake add-rewards test2 100uband --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4

# lock
bandd tx restake lock-power test 90 --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4
bandd tx restake lock-power test2 100 --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4

# add-rewards
bandd tx restake add-rewards test2 80uband --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4

# claim
bandd tx restake claim-rewards --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4
bandd tx restake claim-rewards --from validator --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4

# deactivate
bandd tx restake deactivate test2 --from requester --keyring-backend test --gas-prices 0.0025uband -y 
sleep 4


# undelegate
bandd tx staking unbond bandvaloper1p40yh3zkmhcv0ecqp3mcazy83sa57rgjde6wec 100uband --from requester --keyring-backend test --gas-prices 0.0025uband -y --gas 300000
