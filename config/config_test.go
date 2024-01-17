package config

import (
	"os"
	"testing"
	"time"

	"github.com/prebid/prebid-cache/utils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	v := viper.New()

	setConfigDefaults(v)

	cfg := Configuration{}
	err := v.Unmarshal(&cfg)
	assert.NoError(t, err, "Failed to unmarshal config: %v", err)

	expectedConfig := getExpectedDefaultConfig()
	assert.Equal(t, expectedConfig, cfg, "Expected Configuration instance does not match.")
}

func TestEnvConfig(t *testing.T) {
	defer setEnvVar(t, "PBC_METRICS_INFLUX_HOST", "env-var-defined-metrics-host")()

	// Inside NewConfig() metrics.influx.host sets the default value to ""
	// "config/configtest/sample_full_config.yaml", sets it to  "metrics-host"
	cfg := NewConfig("sample_full_config")

	// assert env variable value supercedes them both
	assert.Equal(t, "env-var-defined-metrics-host", string(cfg.Metrics.Influx.Host), "metrics.influx.host did not equal expected")
}

func TestPrometheusTimeoutDuration(t *testing.T) {
	prometheusConfig := &PrometheusMetrics{
		TimeoutMillisRaw: 5,
	}

	expectedTimeout := time.Duration(5 * 1000 * 1000)
	actualTimeout := prometheusConfig.Timeout()
	assert.Equal(t, expectedTimeout, actualTimeout)
}

// setEnvVar sets an environment variable to a certain value, and returns a function which resets it to its original value.
func setEnvVar(t *testing.T, key string, val string) func() {
	orig, set := os.LookupEnv(key)
	err := os.Setenv(key, val)
	if err != nil {
		t.Fatalf("Error setting evnvironment %s", key)
	}
	if set {
		return func() {
			if os.Setenv(key, orig) != nil {
				t.Fatalf("Error unsetting evnvironment %s", key)
			}
		}
	} else {
		return func() {
			if os.Unsetenv(key) != nil {
				t.Fatalf("Error unsetting evnvironment %s", key)
			}
		}
	}
}

func getExpectedDefaultConfig() Configuration {
	return Configuration{
		Port:          2424,
		AdminPort:     2525,
		IndexResponse: "This application stores short-term data for use in Prebid.",
		Log: Log{
			Level: Info,
		},
		Backend: Backend{
			Type: BackendMemory,
			Memcache: Memcache{
				Hosts: []string{},
			},
			Aerospike: Aerospike{
				Hosts:          []string{},
				MaxReadRetries: 2,
			},
			Cassandra: Cassandra{
				DefaultTTL: utils.CASSANDRA_DEFAULT_TTL_SECONDS,
			},
			Redis: Redis{
				ExpirationMinutes: utils.REDIS_DEFAULT_EXPIRATION_MINUTES,
			},
		},
		Compression: Compression{
			Type: CompressionType("snappy"),
		},
		RateLimiting: RateLimiting{
			Enabled:              true,
			MaxRequestsPerSecond: 100,
		},
		RequestLimits: RequestLimits{
			MaxSize:       10240,
			MaxNumValues:  10,
			MaxTTLSeconds: 3600,
		},
		Routes: Routes{
			AllowPublicWrite: true,
		},
	}
}

// Returns a Configuration object that matches the values found in the `sample_full_config.yaml`
func getExpectedFullConfigForTestFile() Configuration {
	return Configuration{
		Port:          9000,
		AdminPort:     2525,
		IndexResponse: "Any index response",
		Log: Log{
			Level: Info,
		},
		RateLimiting: RateLimiting{
			Enabled:              false,
			MaxRequestsPerSecond: 150,
		},
		RequestLimits: RequestLimits{
			MaxSize:          10240,
			MaxNumValues:     10,
			MaxTTLSeconds:    5000,
			AllowSettingKeys: true,
		},
		Backend: Backend{
			Type: BackendMemory,
			Aerospike: Aerospike{
				DefaultTTLSecs:      3600,
				Host:                "aerospike.prebid.com",
				Hosts:               []string{"aerospike2.prebid.com", "aerospike3.prebid.com"},
				Port:                3000,
				Namespace:           "whatever",
				User:                "foo",
				Password:            "bar",
				MaxReadRetries:      2,
				ConnIdleTimeoutSecs: 2,
			},
			Cassandra: Cassandra{
				Hosts:      "127.0.0.1",
				Keyspace:   "prebid",
				DefaultTTL: 60,
			},
			Memcache: Memcache{
				Hosts: []string{"10.0.0.1:11211", "127.0.0.1"},
			},
			Redis: Redis{
				Host:              "127.0.0.1",
				Port:              6379,
				Password:          "redis-password",
				Db:                1,
				ExpirationMinutes: 1,
				TLS: RedisTLS{
					Enabled:            false,
					InsecureSkipVerify: false,
				},
			},
		},
		Compression: Compression{
			Type: CompressionType("snappy"),
		},
		Metrics: Metrics{
			Type: MetricsType("none"),
			Influx: InfluxMetrics{
				Host:     "metrics-host",
				Database: "metrics-database",
				Username: "metrics-username",
				Password: "metrics-password",
				Enabled:  true,
			},
			Prometheus: PrometheusMetrics{
				Port:             8080,
				Namespace:        "prebid",
				Subsystem:        "cache",
				TimeoutMillisRaw: 100,
				Enabled:          true,
			},
		},
		Routes: Routes{
			AllowPublicWrite: true,
		},
	}
}
