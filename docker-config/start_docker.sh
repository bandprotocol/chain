#!/bin/bash

source ~/.profile
# docker rm --force $(docker ps -a -q)

DIR=`dirname "$0"`

# remove old genesis
rm -rf ~/.band*

make install
make faucet

# initial new node
bandd init node-validator --chain-id bandchain

# add data sources to genesis

# chmod +x $DIR/add_os_ds.sh
# $DIR/add_os_ds.sh

# create acccounts
echo "lock nasty suffer dirt dream fine fall deal curtain plate husband sound tower mom crew crawl guard rack snake before fragile course bacon range" \
    | bandd keys add validator1 --recover --keyring-backend test

echo "loyal damage diet label ability huge dad dash mom design method busy notable cash vast nerve congress drip chunk cheese blur stem dawn fatigue" \
    | bandd keys add validator2 --recover --keyring-backend test

echo "whip desk enemy only canal swear help walnut cannon great arm onion oval doctor twice dish comfort team meat junior blind city mask aware" \
    | bandd keys add validator3 --recover --keyring-backend test

echo "unfair beyond material banner okay genre camera dumb grit balcony permit room intact code degree execute twin flip half salt script cause demand recipe" \
    | bandd keys add validator4 --recover --keyring-backend test

echo "smile stem oven genius cave resource better lunar nasty moon company ridge brass rather supply used horn three panic put venue analyst leader comic" \
    | bandd keys add requester --recover --keyring-backend test

echo "audit silver absorb involve more aspect girl report open gather excite mirror bar hammer clay tackle negative example gym group finger shop stool seminar" \
    | bandd keys add relayer --recover --keyring-backend test

echo "erase relief tree tobacco around knee concert toast diesel melody rule sight forum camera oil sick leopard valid furnace casino post dumb tag young" \
    | bandd keys add account1 --recover --keyring-backend test

echo "thought insane behind cool expand clarify strategy occur arrive broccoli middle despair foot cake genuine dawn goose abuse curve identify dinner derive genre effort" \
    | bandd keys add account2 --recover --keyring-backend test

echo "drop video mention casual soldier ostrich resemble harvest casual step design gasp grunt lab meadow buzz envelope today spy cliff column habit fall eyebrow" \
    | bandd keys add account3 --recover --keyring-backend test

# add accounts to genesis
bandd genesis add-genesis-account validator1 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account validator2 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account validator3 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account validator4 10000000000000uband --keyring-backend test
bandd genesis add-genesis-account requester 100000000000000uband --keyring-backend test
bandd genesis add-genesis-account relayer 100000000000000uband --keyring-backend test
bandd genesis add-genesis-account account1 100000000000000uband --keyring-backend test
bandd genesis add-genesis-account account2 100000000000000uband --keyring-backend test
bandd genesis add-genesis-account account3 100000000000000uband --keyring-backend test

# create copy of config.toml
cp ~/.band/config/config.toml ~/.band/config/config.toml.temp
cp -r ~/.band/files docker-config/

# modify moniker
sed 's/node-validator/🙎‍♀️Alice \& Co./g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

# register initial validators
bandd genesis gentx validator1 100000000uband \
    --chain-id bandchain \
    --node-id 11392b605378063b1c505c0ab123f04bd710d7d7 \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/V/OZek6B2PMh6XEJJ+IsLm0w+22PdJqeSgevs7O3kJ"}' \
    --details "Alice's Adventures in Wonderland (commonly shortened to Alice in Wonderland) is an 1865 novel written by English author Charles Lutwidge Dodgson under the pseudonym Lewis Carroll." \
    --website "https://www.alice.org/" \
    --ip multi-validator1-node \
    --keyring-backend test

# modify moniker
sed 's/node-validator/Bobby.fish 🐡/g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

bandd genesis gentx validator2 100000000uband \
    --chain-id bandchain \
    --node-id 0851086afcd835d5a6fb0ffbf96fcdf74fec742e \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AnJK4pz+t0lwUdCe39joIjUsTINht1dkdkW3jIzHTOiF"}' \
    --details "Fish is best known for his appearances with Ring of Honor (ROH) from 2013 to 2017, where he wrestled as one-half of the tag team reDRagon and held the ROH World Tag Team Championship three times and the ROH World Television Championship once." \
    --website "https://www.wwe.com/superstars/bobby-fish" \
    --ip multi-validator2-node \
    --keyring-backend test

# modify moniker
sed 's/node-validator/Carol/g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

bandd genesis gentx validator3 100000000uband \
    --chain-id bandchain \
    --node-id 7b58b086dd915a79836eb8bfa956aeb9488d13b0 \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A6VP+qhMjy95h4Lei5YqhHhOKISHp0eBOghXJDpg4roz"}' \
    --details "Carol Susan Jane Danvers is a fictional superhero appearing in American comic books published by Marvel Comics. Created by writer Roy Thomas and artist Gene Colan." \
    --website "https://www.marvel.com/characters/captain-marvel-carol-danvers" \
    --ip multi-validator3-node \
    --keyring-backend test

# modify moniker
sed 's/node-validator/Eve 🦹🏿‍♂️the evil with a really long moniker name/g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

bandd genesis gentx validator4 100000000uband \
    --chain-id bandchain \
    --node-id 63808bd64f2ec19acb2a494c8ce8467c595f6fba \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A9A3CPFh0Vg/SeQmCkKysI07oYbXgDojzDrNEvB02ddv"}' \
    --details "Evil is an American supernatural drama television series created by Robert King and Michelle King that premiered on September 26, 2019, on CBS. The series is produced by CBS Television Studios and King Size Productions." \
    --website "https://www.imdb.com/title/tt9055008/" \
    --ip multi-validator4-node \
    --keyring-backend test

# remove temp test
rm -rf ~/.band/config/config.toml.temp

# collect genesis transactions
bandd genesis collect-gentxs

# copy genesis to the proper location!
cp ~/.band/config/genesis.json $DIR/genesis.json
cat <<< $(jq '.app_state.gov.params.voting_period = "60s"' $DIR/genesis.json) > $DIR/genesis.json
cat <<< $(jq '.app_state.feeds.params.current_feeds_update_interval = "10"' $DIR/genesis.json) > $DIR/genesis.json
cat <<< $(jq '.app_state.bandtss.params.min_transition_duration = "60s"' $DIR/genesis.json) > $DIR/genesis.json
cat <<< $(jq '.app_state.bandtss.params.inactive_penalty_duration = "60s"' $DIR/genesis.json) > $DIR/genesis.json
cat <<< $(jq '.app_state.oracle.params.inactive_penalty_duration = "60000000000"' $DIR/genesis.json) > $DIR/genesis.json
cat <<< $(jq --arg addr "$(bandd keys show requester -a --keyring-backend test)" '.app_state.feeds.params.admin = $addr' $DIR/genesis.json) > $DIR/genesis.json
cat <<< $(jq '.app_state.restake.params.allowed_denoms = ["uband"]' $DIR/genesis.json) > $DIR/genesis.json

# Build
docker-compose up -d --build --force-recreate

sleep 10

for v in {1..4}
do
    rm -rf ~/.yoda3
    yoda3 config chain-id bandchain
    yoda3 config node tcp://multi-validator$v-node:26657
    yoda3 config validator $(bandd keys show validator$v -a --bech val --keyring-backend test)
    yoda3 config executor "rest:https://asia-southeast2-band-playground.cloudfunctions.net/test-runtime-executor?timeout=10s"

    # activate validator
    echo "y" | bandd tx oracle activate --from validator$v --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for activation transaction success
    sleep 4

    for i in $(eval echo {1..4})
    do
        # add reporter key
        yoda3 keys add reporter$i
    done

    # send band tokens to reporters
    echo "y" | bandd tx bank multi-send validator$v $(yoda3 keys list -a) 1000000uband --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for sending band tokens transaction success
    sleep 4

    # add reporter to bandchain
    echo "y" | bandd tx oracle add-reporters $(yoda3 keys list -a) --from validator$v --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for adding reporter transaction success
    sleep 4

    docker create --network chain_bandchain --name bandchain-yoda${v} band-validator:latest yoda3 r
    docker cp ~/.yoda3 bandchain-yoda${v}:/root/.yoda3
    docker start bandchain-yoda${v}
done

sleep 10

for v in {1..3}
do
    rm -rf ~/.cylinder
    cylinder config node tcp://multi-validator$v-node:26657
    cylinder config chain-id bandchain
    cylinder config granter $(bandd keys show account$v -a --keyring-backend test)
    cylinder config max-messages 10
    cylinder config broadcast-timeout "5m"
    cylinder config rpc-poll-interval "1s"
    cylinder config max-try 5
    cylinder config gas-prices "0uband"
    cylinder config min-de 100
    cylinder config gas-adjust-start 1.6
    cylinder config gas-adjust-step 0.2
    cylinder config random-secret "$(openssl rand -hex 32)"
    cylinder config checking-de-interval "1m"

    for i in $(eval echo {1..4})
    do
        # add signer key
        cylinder keys add signer$i
    done

    # send band tokens to grantees
    echo "y" | bandd tx bank multi-send account$v $(cylinder keys list -a) 1000000uband --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync
    sleep 4

    # grant tss
    echo "y" | bandd tx tss add-grantees $(cylinder keys list -a) --from account$v --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband --gas 700000 -b sync
    sleep 4

    docker create --network chain_bandchain --name bandchain-cylinder${v} band-validator:latest cylinder run
    docker cp ~/.cylinder bandchain-cylinder${v}:/root/.cylinder
    docker start bandchain-cylinder${v}
done

# pull latest image first
docker pull bandprotocol/bothan-api:latest

for v in {1..4}
do
    # run bothan image
    docker run --log-opt max-size=10m --log-opt max-file=3 --network chain_bandchain -d --name bothan$v -v "$(pwd)/docker-config/bothan-config.toml:/root/.bothan/config.toml" bandprotocol/bothan-api:latest

    rm -rf ~/.grogu
    grogu config chain-id bandchain
    grogu config nodes "tcp://multi-validator$v-node:26657"
    grogu config validator $(bandd keys show validator$v -a --bech val --keyring-backend test)

    # change url to bothan image
    grogu config bothan "bothan$v:50051"

    # activate validator
    echo "y" | bandd tx oracle activate --from validator$v --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for activation transaction success
    sleep 4

    for i in $(eval echo {1..4})
    do
        # add feeder key
        grogu keys add feeder$i
    done

    # send band tokens to feeders
    echo "y" | bandd tx bank multi-send validator$v  $(grogu keys list -a) 1000000uband --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for sending band tokens transaction success
    sleep 4

    # add feeder to bandchain
    echo "y" | bandd tx feeds add-feeders $(grogu keys list -a) --from validator$v --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for adding feeder transaction success
    sleep 4

    docker create --network chain_bandchain --name bandchain-grogu${v} band-validator:latest grogu r
    docker cp ~/.grogu bandchain-grogu${v}:/root/.grogu
    docker start bandchain-grogu${v}
done


# Create faucet container
rm -rf ~/.faucet
faucet config chain-id bandchain
faucet config node tcp://query-node:26657
faucet config port 5005
faucet config gas-prices "0.0025uband"
for i in $(eval echo {1..5})
do
    # add worker key
    faucet keys add worker$i

    # send band tokens to worker
    echo "y" | bandd tx bank send requester $(faucet keys show worker$i) 1000000000000uband --keyring-backend test --chain-id bandchain --gas-prices 0.0025uband -b sync

    # wait for adding token transaction success
    sleep 4
done

docker create --network chain_bandchain --name bandchain-faucet -p 5005:5005 band-validator:latest faucet r
docker cp ~/.faucet bandchain-faucet:/root/.faucet
docker start bandchain-faucet
