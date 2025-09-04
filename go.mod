module github.com/dboxed/dboxed

go 1.24.3

//replace github.com/dboxed/dboxed-common => ../dboxed-common

//replace github.com/dboxed/dboxed-volume => ../dboxed-volume

require (
	github.com/alecthomas/kong v1.12.1
	github.com/compose-spec/compose-go/v2 v2.7.1
	github.com/cyphar/filepath-securejoin v0.4.1
	github.com/dboxed/dboxed-common v0.0.0-20250902144135-ea47c42c7a59
	github.com/dboxed/dboxed-volume v0.0.0-20250903150732-bce1c4eb0016
	github.com/dgraph-io/badger/v4 v4.8.0
	github.com/dustin/go-humanize v1.0.1
	github.com/fsnotify/fsnotify v1.9.0
	github.com/gofrs/flock v0.8.1
	github.com/google/go-containerregistry v0.20.6
	github.com/klauspost/compress v1.18.0
	github.com/miekg/dns v1.1.66
	github.com/moby/go-archive v0.1.0
	github.com/nats-io/nats.go v1.43.0
	github.com/nats-io/nkeys v0.4.11
	github.com/opencontainers/image-spec v1.1.1
	github.com/opencontainers/runtime-spec v1.2.1
	github.com/vishvananda/netlink v1.3.0
	github.com/vishvananda/netns v0.0.4
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba
	golang.org/x/sys v0.34.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	sigs.k8s.io/yaml v1.5.0
)

require (
	github.com/bytedance/sonic v1.13.3 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/coreos/go-oidc/v3 v3.15.0 // indirect
	github.com/danielgtaylor/huma/v2 v2.34.1 // indirect
	github.com/dgraph-io/ristretto/v2 v2.2.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v28.3.1+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.3 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/gin-gonic/gin v1.10.1 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/vbatts/tar-split v0.12.1 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/arch v0.18.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
