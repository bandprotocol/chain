git checkout v2.5.4
make install

./scripts/generate_genesis.sh
./scripts/test_script/setup_voting_period.sh 60s

sed -i -e \
    "s/^minimum-gas-prices *=.*/minimum-gas-prices = \"0.0025uband\"/" \
    ~/.band/config/app.toml

sed -i -e \
  '/\[api\]/,+10 s/enable = .*/enable = true/' \
  ~/.band/config/app.toml

sed -i -e \
  '/\[mempool\]/,+10 s/version = .*/version = \"v1\"/' \
  ~/.band/config/config.toml

./scripts/test_script/setup_chains.sh /Users/ongart/Development/band/chain v2.5.4 cosmos-sdk-v0.47 v2_6

./scripts/test_script/start_cosmovisor.sh

# then run the open_vote_proposal_25.sh
