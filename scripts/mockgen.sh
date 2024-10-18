#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/oracle/types/expected_keepers.go -package testutil -destination x/oracle/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/feeds/types/expected_keepers.go -package testutil -destination x/feeds/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/restake/types/expected_keepers.go -package testutil -destination x/restake/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/bandtss/types/expected_keepers.go -package testutil -destination x/bandtss/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/tss/types/expected_keepers.go -package testutil -destination x/tss/testutil/expected_keepers_mocks.go

$mockgen_cmd -source=grogu/submitter/expected_types.go -package testutil -destination grogu/submitter/testutil/expected_types_mock.go
$mockgen_cmd -source=grogu/signaller/expected_types.go -package testutil -destination grogu/signaller/testutil/expected_types_mock.go
$mockgen_cmd -source=grogu/updater/expected_types.go -package testutil -destination grogu/updater/testutil/expected_types_mock.go
