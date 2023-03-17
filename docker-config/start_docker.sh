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

# add accounts to genesis
bandd add-genesis-account validator1 10000000000000uband --keyring-backend test
bandd add-genesis-account validator2 10000000000000uband --keyring-backend test
bandd add-genesis-account validator3 10000000000000uband --keyring-backend test
bandd add-genesis-account validator4 10000000000000uband --keyring-backend test
bandd add-genesis-account requester 100000000000000uband --keyring-backend test
bandd add-genesis-account relayer 100000000000000uband --keyring-backend test

# create copy of config.toml
cp ~/.band/config/config.toml ~/.band/config/config.toml.temp
cp -r ~/.band/files docker-config/

# modify moniker
sed 's/node-validator/🙎‍♀️Alice \& Co./g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

# register initial validators
bandd gentx validator1 100000000uband \
    --chain-id bandchain \
    --node-id 11392b605378063b1c505c0ab123f04bd710d7d7 \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/V/OZek6B2PMh6XEJJ+IsLm0w+22PdJqeSgevs7O3kJ"}' \
    --details "Alice's Adventures in Wonderland (commonly shortened to Alice in Wonderland) is an 1865 novel written by English author Charles Lutwidge Dodgson under the pseudonym Lewis Carroll." \
    --website "https://www.alice.org/" \
    --ip multi-validator1-node \
    --keyring-backend test

# modify moniker
sed 's/node-validator/Bobby.fish 🐡/g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

bandd gentx validator2 100000000uband \
    --chain-id bandchain \
    --node-id 0851086afcd835d5a6fb0ffbf96fcdf74fec742e \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AnJK4pz+t0lwUdCe39joIjUsTINht1dkdkW3jIzHTOiF"}' \
    --details "Fish is best known for his appearances with Ring of Honor (ROH) from 2013 to 2017, where he wrestled as one-half of the tag team reDRagon and held the ROH World Tag Team Championship three times and the ROH World Television Championship once." \
    --website "https://www.wwe.com/superstars/bobby-fish" \
    --ip multi-validator2-node \
    --keyring-backend test

# modify moniker
sed 's/node-validator/Carol/g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

bandd gentx validator3 100000000uband \
    --chain-id bandchain \
    --node-id 7b58b086dd915a79836eb8bfa956aeb9488d13b0 \
    --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A6VP+qhMjy95h4Lei5YqhHhOKISHp0eBOghXJDpg4roz"}' \
    --details "Carol Susan Jane Danvers is a fictional superhero appearing in American comic books published by Marvel Comics. Created by writer Roy Thomas and artist Gene Colan." \
    --website "https://www.marvel.com/characters/captain-marvel-carol-danvers" \
    --ip multi-validator3-node \
    --keyring-backend test

# modify moniker
sed 's/node-validator/Eve 🦹🏿‍♂️the evil with a really long moniker name/g' ~/.band/config/config.toml.temp > ~/.band/config/config.toml

bandd gentx validator4 100000000uband \
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
bandd collect-gentxs

# copy genesis to the proper location!
cp ~/.band/config/genesis.json $DIR/genesis.json
sed -i -e 's/\"allow_messages\":.*/\"allow_messages\": [\"\/cosmos.authz.v1beta1.MsgExec\", \"\/cosmos.authz.v1beta1.MsgGrant\", \"\/cosmos.authz.v1beta1.MsgRevoke\", \"\/cosmos.bank.v1beta1.MsgSend\", \"\/cosmos.bank.v1beta1.MsgMultiSend\", \"\/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"\/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission\", \"\/cosmos.distribution.v1beta1.MsgFundCommunityPool\", \"\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"\/cosmos.feegrant.v1beta1.MsgGrantAllowance\", \"\/cosmos.feegrant.v1beta1.MsgRevokeAllowance\", \"\/cosmos.gov.v1beta1.MsgVoteWeighted\", \"\/cosmos.gov.v1beta1.MsgSubmitProposal\", \"\/cosmos.gov.v1beta1.MsgDeposit\", \"\/cosmos.gov.v1beta1.MsgVote\", \"\/cosmos.staking.v1beta1.MsgEditValidator\", \"\/cosmos.staking.v1beta1.MsgDelegate\", \"\/cosmos.staking.v1beta1.MsgUndelegate\", \"\/cosmos.staking.v1beta1.MsgBeginRedelegate\", \"\/cosmos.staking.v1beta1.MsgCreateValidator\", \"\/cosmos.vesting.v1beta1.MsgCreateVestingAccount\", \"\/ibc.applications.transfer.v1.MsgTransfer\"]/g' $DIR/genesis.json

# Build
docker-compose up -d --build

sleep 10

for v in {1..4}
do
    rm -rf ~/.yoda
    yoda config chain-id bandchain
    yoda config node tcp://multi-validator$v-node:26657
    yoda config validator $(bandd keys show validator$v -a --bech val --keyring-backend test)
    yoda config executor "rest:https://asia-southeast2-band-playground.cloudfunctions.net/test-runtime-executor?timeout=10s"

    # activate validator
    echo "y" | bandd tx oracle activate --from validator$v --keyring-backend test --chain-id bandchain -b block

    # wait for activation transaction success
    sleep 2

    for i in $(eval echo {1..1})
    do
        # add reporter key
        yoda keys add reporter$i
    done

    # send band tokens to reporters
    echo "y" | bandd tx bank send validator$v  $(yoda keys list -a) 1000000uband --keyring-backend test --chain-id bandchain -b block

    # wait for sending band tokens transaction success
    sleep 2

    # add reporter to bandchain
    echo "y" | bandd tx oracle add-reporters $(yoda keys list -a) --from validator$v --keyring-backend test --chain-id bandchain -b block

    # wait for addding reporter transaction success
    sleep 2

    docker create --network chain_bandchain --name bandchain_oracle${v} band-validator:latest yoda r
    docker cp ~/.yoda bandchain_oracle${v}:/root/.yoda
    docker start bandchain_oracle${v}
done

# Create faucet container
rm -rf ~/.faucet
faucet config chain-id bandchain
faucet config node tcp://query-node:26657
faucet config port 5005
for i in $(eval echo {1..5})
do
    # add worker key
    faucet keys add worker$i

    # send band tokens to worker
    echo "y" | bandd tx bank send requester $(faucet keys show worker$i) 1000000000000uband --keyring-backend test --chain-id bandchain -b block

    # wait for addding reporter transaction success
    sleep 2
done

docker create --network chain_bandchain --name bandchain_faucet -p 5005:5005 band-validator:latest faucet r
docker cp ~/.faucet bandchain_faucet:/root/.faucet
docker start bandchain_faucet
