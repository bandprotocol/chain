#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/oracle/types/expected_keepers.go -package testutil -destination x/oracle/testutil/expected_keepers_mocks.go