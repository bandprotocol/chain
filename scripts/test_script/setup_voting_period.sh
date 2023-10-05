#!/bin/bash

new_voting_period=$1
genesis_file="$HOME/.band/config/genesis.json"

# update voting period for testing
voting_period_value=$(jq -r '.app_state.gov.params.voting_period' $genesis_file)
if [ "$voting_period_value" != "null" ]; then
    cat <<< $(jq --arg new_voting_period "$new_voting_period" '.app_state.gov.params.voting_period = $new_voting_period' $genesis_file) > $genesis_file
fi

voting_period_value=$(jq -r '.app_state.gov.voting_params.voting_period' $genesis_file)
if [ "$voting_period_value" != "null" ]; then
    cat <<< $(jq --arg new_voting_period "$new_voting_period" '.app_state.gov.voting_params.voting_period = $new_voting_period' $genesis_file) > $genesis_file
fi

voting_period_value=$(jq -r '.app_state.council.params.voting_period' $genesis_file)
if [ "$voting_period_value" != "null" ]; then
    cat <<< $(jq --arg new_voting_period "$new_voting_period" '.app_state.council.params.voting_period = $new_voting_period' $genesis_file) > $genesis_file
fi
