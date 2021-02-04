DIR=`dirname "$0"`

rm -rf ~/.band

# initial new node
bandd init validator --chain-id bandchain
echo "lock nasty suffer dirt dream fine fall deal curtain plate husband sound tower mom crew crawl guard rack snake before fragile course bacon range" \
    | bandd keys add validator --recover --keyring-backend test
echo "smile stem oven genius cave resource better lunar nasty moon company ridge brass rather supply used horn three panic put venue analyst leader comic" \
    | bandd keys add requester --recover --keyring-backend test

cp ./docker-config/single-validator/priv_validator_key.json ~/.band/config/priv_validator_key.json
cp ./docker-config/single-validator/node_key.json ~/.band/config/node_key.json

# add accounts to genesis
bandd add-genesis-account validator 10000000000000uband --keyring-backend test
bandd add-genesis-account requester 10000000000000uband --keyring-backend test


# register initial validators
bandd gentx validator 100000000uband \
    --chain-id bandchain \
    --keyring-backend test

# collect genesis transactions
bandd collect-gentxs


