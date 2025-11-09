module github.com/df-mc/dragonfly

go 1.24.0

toolchain go1.24.4

require (
	github.com/brentp/intintmap v0.0.0-20190211203843-30dc0ade9af9
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/df-mc/goleveldb v1.1.9
	github.com/df-mc/worldupgrader v1.0.20
	github.com/go-gl/mathgl v1.2.0
	github.com/google/uuid v1.6.0
	github.com/pelletier/go-toml v1.9.5
	github.com/sandertv/gophertunnel v1.51.0
	github.com/segmentio/fasthash v1.0.3
	golang.org/x/exp v0.0.0-20250103183323-7d7fa50e5329
	golang.org/x/mod v0.25.0
	golang.org/x/text v0.27.0
	golang.org/x/tools v0.34.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v2 v2.3.0
)

require (
	github.com/df-mc/jsonc v1.0.5 // indirect
	github.com/go-jose/go-jose/v4 v4.1.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/sandertv/go-raknet v1.14.3-0.20250305181847-6af3e95113d6 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)

replace github.com/df-mc/dragonfly => ./
