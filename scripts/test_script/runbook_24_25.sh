git checkout v2.4.1
make install

./scripts/generate_genesis.sh
./scripts/test_script/setup_voting_period.sh 60s

./scripts/test_script/setup_chains.sh /Users/ongart/Development/band/chain v2.4.1 v2.5.4 v2_5

./scripts/test_script/start_cosmovisor.sh

# then run the open_vote_proposal_25.sh
