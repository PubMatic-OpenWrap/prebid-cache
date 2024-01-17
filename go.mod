module github.com/PubMatic-OpenWrap/prebid-cache

go 1.19

replace github.com/bradfitz/gomemcache => github.com/google/gomemcache v0.0.0-20200326162346-94281991662a

replace github.com/prebid/prebid-cache => ./

require (
	git.pubmatic.com/PubMatic/go-common v0.0.0-20231211162912-5d6b67bde771
	github.com/aerospike/aerospike-client-go/v6 v6.7.0
	github.com/didip/tollbooth/v6 v6.1.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gocql/gocql v1.0.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang/glog v1.0.0
	github.com/golang/snappy v0.0.4
	github.com/google/gomemcache v0.0.0-20210709172713-c1c93e4523ee
	github.com/julienschmidt/httprouter v1.3.0
	github.com/prebid/prebid-cache v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.12.2
	github.com/prometheus/client_model v0.2.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rs/cors v1.8.2
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.2
	github.com/vrischmann/go-metrics-influxdb v0.1.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-pkgz/expirable-cache v0.0.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/yuin/gopher-lua v0.0.0-20220504180219-658193537a64 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
