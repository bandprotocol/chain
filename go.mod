module github.com/bandprotocol/chain

go 1.15

require (
	github.com/bandprotocol/go-owasm v0.0.0-20210311072328-a6859c27139c
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/go-sql-driver/mysql v1.4.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.4.3
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/lib/pq v1.9.0
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/oasisprotocol/oasis-core/go v0.0.0-20200730171716-3be2b460b3ac
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/poy/onpar v1.1.2 // indirect
	github.com/prometheus/client_golang v1.8.0
	github.com/rakyll/statik v0.1.7
	github.com/segmentio/kafka-go v0.3.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/tendermint/tendermint v0.34.9
	github.com/tendermint/tm-db v0.6.4
	github.com/ziutek/mymysql v1.5.4 // indirect
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
