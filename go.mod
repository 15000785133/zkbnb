module github.com/zecrey-labs/zecrey-legend

go 1.16

require (
	github.com/zeromicro/go-zero v1.3.3
	gorm.io/gorm v1.23.4
)

require (
	github.com/consensys/gnark v0.7.0
	github.com/consensys/gnark-crypto v0.7.0
	github.com/eko/gocache/v2 v2.3.1
	github.com/ethereum/go-ethereum v1.10.17
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.3.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/zecrey-labs/zecrey-crypto v0.0.5-legend
	github.com/zecrey-labs/zecrey-eth-rpc v0.0.14
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
	gorm.io/driver/postgres v1.3.6
	k8s.io/client-go v0.24.1 // indirect
)

//replace github.com/zecrey-labs/zecrey-crypto => ../zecrey-crypto
//
//replace github.com/zecrey-labs/zecrey-eth-rpc => ../zecrey-eth-rpc
