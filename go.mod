module g.tesamc.com/IT/zaipkg

go 1.16

require (
	g.tesamc.com/IT/zproto v0.0.0
	github.com/BurntSushi/toml v0.3.1
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/docker/go-units v0.4.0
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/elastic/go-hdrhistogram v0.1.0
	github.com/go-redis/redis/v8 v8.11.2
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.2.0
	github.com/gyuho/linux-inspect v0.0.0-20180929231013-a492bfc5f12a
	github.com/jaypipes/ghw v0.8.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kr/pretty v0.2.1 // indirect
	github.com/lni/goutils v1.2.0
	github.com/panjf2000/ants/v2 v2.4.6
	github.com/pierrec/lz4/v4 v4.1.8
	github.com/prometheus/client_golang v1.6.0
	github.com/spf13/cast v1.3.1
	github.com/stretchr/testify v1.6.1
	github.com/templexxx/cpu v0.0.8-0.20210423085042-1c810926b5dd
	github.com/templexxx/fnc v1.0.1
	github.com/templexxx/tsc v1.0.1
	github.com/urfave/negroni/v2 v2.0.2
	github.com/zaibyte/nanozap v0.0.7
	github.com/zeebo/xxh3 v1.0.0-rc3.0.20210921232450-c77878a38204
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20210514084401-e8d321eab015
	google.golang.org/grpc v1.29.1
)

// TODO GitLAB proxy issues
replace g.tesamc.com/IT/zproto v0.0.0 => ../zproto
