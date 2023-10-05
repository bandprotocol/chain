#!/bin/bash

# make new version binary
cd $1
git checkout $3
make install

# setup cosmovisor for upgrade version
mkdir -p $HOME/.band/cosmovisor/upgrades/$4/bin
cp $HOME/go/bin/bandd $HOME/.band/cosmovisor/upgrades/$4/bin

# make old version binary
git checkout $2
make install

# setup cosmovisor for genesis version 
mkdir -p $HOME/.band/cosmovisor/genesis/bin
mkdir -p $HOME/.band/cosmovisor/upgrades
cp $HOME/go/bin/bandd $HOME/.band/cosmovisor/genesis/bin

