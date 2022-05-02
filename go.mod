module github.com/PubMatic-OpenWrap/prebid-cache

go 1.16

replace github.com/bradfitz/gomemcache => github.com/google/gomemcache v0.0.0-20200326162346-94281991662a

replace github.com/prebid/prebid-cache => ./

require (
	git.pubmatic.com/PubMatic/go-common.git v0.0.0-20211116062746-840b999f668b
	github.com/aerospike/aerospike-client-go v4.0.0+incompatible
	github.com/bradfitz/gomemcache v0.0.0-20180710155616-bc664df96737
	github.com/didip/tollbooth v2.2.0+incompatible
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-redis/redis v6.12.1-0.20180718122851-ee41b9092371+incompatible
	github.com/gocql/gocql v0.0.0-20180617115710-e06f8c1bcd78
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang/glog v1.0.0
	github.com/golang/snappy v0.0.4
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/influxdata/influxdb v1.9.6 // indirect
	github.com/julienschmidt/httprouter v1.3.0
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/prebid/prebid-cache v0.0.0-20220302220808-955b74c1a07e
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/client_model v0.2.0
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/rs/cors v1.6.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.0.2
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/go-metrics-influxdb v0.0.0-20160917065939-43af8332c303
	github.com/yuin/gopher-lua v0.0.0-20220413183635-c841877397d8 // indirect
)
