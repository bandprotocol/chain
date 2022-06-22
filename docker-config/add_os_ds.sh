#!/bin/bash

# Set `DIR` to your path of genesis directory.
DIR=~/genesis_ds_os/genesis

bandd add-data-source \
    "CoinGecko Cryptocurrency Price" \
    "Retrieves current price of a cryptocurrency from https://www.coingecko.com" \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs \
    $DIR/datasources/coingecko_price.py \
    0uband \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs

bandd add-data-source \
    "CryptoCompare Cryptocurrency Price" \
    "Retrieves current price of a cryptocurrency from https://www.cryptocompare.com" \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs \
    $DIR/datasources/cryptocompare_price.py \
    0uband \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs

bandd add-data-source \
    "Binance Cryptocurrency Price" \
    "Retrieves current price of a cryptocurrency from https://www.binance.com/en" \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs \
    $DIR/datasources/binance_price.py \
    0uband \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs

bandd add-oracle-script \
    "Cryptocurrency Price in USD" \
    "Oracle script that queries the average cryptocurrency price using current price data from CoinGecko, CryptoCompare, and Binance" \
    "{symbol:string,multiplier:u64}/{px:u64}" \
    "https://ipfs.io/ipfs/QmQqxHLszpbCy8Hk2ame3pPAxUUAyStBrVdGdDgrfAngAv" \
    band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs \
    $DIR/res/crypto_price.wasm
