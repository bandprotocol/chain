DIR=`dirname "$0"`

rm -rf ~/.band1
rm -rf ~/.band2

make install

# initial new node
bandd init validator1 --chain-id bandchain --home ~/.band1
bandd init validator2 --chain-id bandchain --home ~/.band2

# validator1
echo "lock nasty suffer dirt dream fine fall deal curtain plate husband sound tower mom crew crawl guard rack snake before fragile course bacon range" \
    | bandd keys add validator1 --recover --keyring-backend test --home ~/.band1
echo "lock nasty suffer dirt dream fine fall deal curtain plate husband sound tower mom crew crawl guard rack snake before fragile course bacon range" \
    | bandd keys add validator1 --recover --keyring-backend test --home ~/.band2

# validator2
echo "loyal damage diet label ability huge dad dash mom design method busy notable cash vast nerve congress drip chunk cheese blur stem dawn fatigue" \
    | bandd keys add validator2 --recover --keyring-backend test --home ~/.band1
echo "loyal damage diet label ability huge dad dash mom design method busy notable cash vast nerve congress drip chunk cheese blur stem dawn fatigue" \
    | bandd keys add validator2 --recover --keyring-backend test --home ~/.band2

# requester
echo "smile stem oven genius cave resource better lunar nasty moon company ridge brass rather supply used horn three panic put venue analyst leader comic" \
    | bandd keys add requester --recover --keyring-backend test --home ~/.band1
echo "smile stem oven genius cave resource better lunar nasty moon company ridge brass rather supply used horn three panic put venue analyst leader comic" \
    | bandd keys add requester --recover --keyring-backend test --home ~/.band2

node1=$(bandd tendermint show-node-id --home ~/.band1)
pub1=$(bandd tendermint show-validator --home ~/.band1)

node2=$(bandd tendermint show-node-id --home ~/.band2)
pub2=$(bandd tendermint show-validator --home ~/.band2)

# add accounts to genesis
bandd add-genesis-account validator1 10000000000000uband --keyring-backend test --home ~/.band1
bandd add-genesis-account validator2 10000000000000uband --keyring-backend test --home ~/.band1
bandd add-genesis-account requester 10000000000000uband --keyring-backend test --home ~/.band1

# register initial validators
bandd gentx validator1 100000000uband \
    --chain-id bandchain \
    --node-id $node1 \
    --pubkey "$pub1" \
    --moniker validator1 \
    --keyring-backend test \
    --home ~/.band1

# register initial validators
bandd gentx validator2 100000000uband \
    --chain-id bandchain \
    --node-id $node2 \
    --pubkey "$pub2" \
    --moniker validator2 \
    --keyring-backend test \
    --home ~/.band1

# collect genesis transactions
bandd collect-gentxs --home ~/.band1

cp ~/.band1/config/genesis.json ~/.band2/config/genesis.json
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:16656"#g' ~/.band2/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:16657"#g' ~/.band2/config/config.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:1316"#g' ~/.band2/config/app.toml
sed -i -e 's#"0.0.0.0:9090"#"0.0.0.0:8090"#g' ~/.band2/config/app.toml
sed -i -e 's#"0.0.0.0:9091"#"0.0.0.0:8091"#g' ~/.band2/config/app.toml
sed -i -e "s/persistent_peers = \".*\"/persistent_peers = \"$node1@127.0.0.1:26656\"/" ~/.band2/config/config.toml

sed -i -e \
  '/\[api\]/,+10 s/enable = .*/enable = true/' \
  ~/.band1/config/app.toml

# # Correct node
# ./scripts/generate_duo.sh           
# bandd start --seed 0 --home ~/.band1

# # Wrong node - Reset without priv
# bandd start --seed 1 --home ~/.band2
# bandd tendermint unsafe-reset-all --home ~/.band2
# bandd start --seed 0 --home ~/.band2

# # Wrong node - Reset with priv
# bandd start --seed 1 --home ~/.band2
# bandd tendermint reset-state --home ~/.band2
# rm -rf ~/.band2/data/application.db
# bandd start --seed 0 --home ~/.band2
