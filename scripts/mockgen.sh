#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/feeds/types/expected_keepers.go -package testutil -destination x/feeds/testutil/mock_expected_keepers.go
$mockgen_cmd -source x/bandtss/types/expected_keepers.go -package testutil -destination x/bandtss/testutil/mock_expected_keepers.go
$mockgen_cmd -source x/tss/types/expected_keepers.go -package testutil -destination x/tss/testutil/mock_expected_keepers.go
