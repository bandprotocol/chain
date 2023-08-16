DIR=`dirname "$0"`

rm -rf ~/.band

# initial new node
bandd init validator --chain-id bandchain
echo "lock nasty suffer dirt dream fine fall deal curtain plate husband sound tower mom crew crawl guard rack snake before fragile course bacon range" \
    | bandd keys add validator --recover --keyring-backend test
echo "smile stem oven genius cave resource better lunar nasty moon company ridge brass rather supply used horn three panic put venue analyst leader comic" \
    | bandd keys add requester --recover --keyring-backend test
echo "drop video mention casual soldier ostrich resemble harvest casual step design gasp grunt lab meadow buzz envelope today spy cliff column habit fall eyebrow" \
    | bandd keys add tss1 --recover --keyring-backend test
echo "enlist electric thumb valve inherit visa ecology trust cake argue forward hidden thing analyst science treat ice lend pumpkin today ticket purchase process pioneer" \
    | bandd keys add tss2 --recover --keyring-backend test
echo "measure fence mail fluid olive cute empower fossil ahead manage snow marble dash citizen tourist skate assist solution bonus spend tip negative try eyebrow" \
    | bandd keys add tss3 --recover --keyring-backend test


# add accounts to genesis
bandd add-genesis-account validator 10000000000000uband --keyring-backend test
bandd add-genesis-account requester 10000000000000uband --keyring-backend test

## add tss accounts to genesis
bandd add-genesis-account tss1 10000000000000uband --keyring-backend test
bandd add-genesis-account tss2 10000000000000uband --keyring-backend test
bandd add-genesis-account tss3 10000000000000uband --keyring-backend test

# register initial validators
bandd gentx validator 100000000uband \
    --chain-id bandchain \
    --keyring-backend test

# collect genesis transactions
bandd collect-gentxs

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
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "60s"' ~/.band/config/genesis.json) > ~/.band/config/genesis.json
