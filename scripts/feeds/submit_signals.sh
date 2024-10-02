# delegate
bandd tx staking delegate bandvaloper1p40yh3zkmhcv0ecqp3mcazy83sa57rgjde6wec 1000000000000uband --from validator --keyring-backend test --gas-prices 0.0025uband -y --chain-id bandchain
bandd tx staking delegate bandvaloper1p40yh3zkmhcv0ecqp3mcazy83sa57rgjde6wec 1000000000000uband --from requester --keyring-backend test --gas-prices 0.0025uband -y --chain-id bandchain
sleep 3

# signal
bandd tx feeds signal CS:BTC-USD,30000000000 CS:USDT-USD,30000000000 --from validator --keyring-backend test --gas-prices 0.0025uband -y --chain-id bandchain
bandd tx feeds signal CS:BTC-USD,30000000000 CS:USDT-USD,29000000000 --from requester --keyring-backend test --gas-prices 0.0025uband -y --chain-id bandchain
