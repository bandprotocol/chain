#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/feeds/types/expected_keepers.go -package testutil -destination x/feeds/testutil/expected_keepers_mocks.go
