#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/feeds/types/expected_keepers.go -package testutil -destination x/feeds/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/restake/types/expected_keepers.go -package testutil -destination x/restake/testutil/expected_keepers_mocks.go
