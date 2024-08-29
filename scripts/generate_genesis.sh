DIR=`dirname "$0"`

rm -rf ~/.band

# initial new node
bandd init validator --chain-id bandchain
echo "lock nasty suffer dirt dream fine fall deal curtain plate husband sound tower mom crew crawl guard rack snake before fragile course bacon range" \
    | bandd keys add validator --recover --keyring-backend test
echo "smile stem oven genius cave resource better lunar nasty moon company ridge brass rather supply used horn three panic put venue analyst leader comic" \
    | bandd keys add requester --recover --keyring-backend test
echo "erase relief tree tobacco around knee concert toast diesel melody rule sight forum camera oil sick leopard valid furnace casino post dumb tag young" \
    | bandd keys add account1 --recover --keyring-backend test
echo "thought insane behind cool expand clarify strategy occur arrive broccoli middle despair foot cake genuine dawn goose abuse curve identify dinner derive genre effort" \
    | bandd keys add account2 --recover --keyring-backend test
echo "drop video mention casual soldier ostrich resemble harvest casual step design gasp grunt lab meadow buzz envelope today spy cliff column habit fall eyebrow" \
    | bandd keys add account3 --recover --keyring-backend test

# add accounts to genesis
bandd genesis add-genesis-account validator 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account requester 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account account1 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account account2 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account account3 10000000000000uband --keyring-backend test

# register initial validators
bandd genesis gentx validator 100000000uband \
    --chain-id bandchain \
    --keyring-backend test

# collect genesis transactions
bandd genesis collect-gentxs

sed -i -e \
    "s/^minimum-gas-prices *=.*/minimum-gas-prices = \"0.0025uband\"/" \
    ~/.band/config/app.toml

sed -i -e \
  '/\[api\]/,+10 s/enable = .*/enable = true/' \
  ~/.band/config/app.toml

sed -i -e \
  '/\[mempool\]/,+10 s/version = .*/version = \"v1\"/' \
  ~/.band/config/config.toml

# update voting period to be 60s for testing
cat <<< $(jq '.app_state.gov.params.voting_period = "60s"' ~/.band/config/genesis.json) > ~/.band/config/genesis.json

# update blocks per feeds update to 10 blocks for testing
cat <<< $(jq '.app_state.feeds.params.current_feeds_update_interval = "10"' ~/.band/config/genesis.json) > ~/.band/config/genesis.json
