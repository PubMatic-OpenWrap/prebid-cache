package config

import (
	"log"
	"strings"
	"time"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/prebid/prebid-cache/metrics/stats"
	"github.com/spf13/viper"

	"github.com/prebid/prebid-cache/utils"
)

func NewConfig(filename string) Configuration {
	v := viper.New()
	setConfigDefaults(v)
	setEnvVarsLookup(v)
	setConfigFilePath(v, filename)

	// Read configuration file
	err := v.ReadInConfig()
	if err != nil {
		// Make sure the configuration file was not defective
		if _, fileNotFound := err.(viper.ConfigFileNotFoundError); fileNotFound {
			// Config file not found. Just log at info level and start Prebid Cache with default values
			logger.Info("Configuration file not detected. Initializing with default values and environment variable overrides.")
		} else {
			// Config file was found but was defective, Either `UnsupportedConfigError` or `ConfigParseError` was thrown
			logger.Fatal("Configuration file could not be read: %v", err)
		}
	}

	cfg := Configuration{}
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Fatal("Failed to unmarshal config: %v", err)
	}

	cfg.Server.ServerName = utils.GetServerName()

	stats.InitStat(&cfg.Stats)

	var logConf logger.LogConf
	logConf.LogLevel = cfg.OWLog.LogLevel
	logConf.LogPath = cfg.OWLog.LogPath
	logConf.LogRotationTime = cfg.OWLog.LogRotationTime
	logConf.MaxLogFiles = cfg.OWLog.MaxLogFiles
	logConf.MaxLogSize = cfg.OWLog.MaxLogSize
	//Initialize logger
	logger.InitGlog(logConf)

	return cfg
}

func setConfigDefaults(v *viper.Viper) {
	v.SetDefault("port", 2424)
	v.SetDefault("admin_port", 2525)
	v.SetDefault("index_response", "This application stores short-term data for use in Prebid.")
	v.SetDefault("status_response", "")
	v.SetDefault("log.level", "info")
	v.SetDefault("backend.type", "memory")
	v.SetDefault("backend.aerospike.host", "")
	v.SetDefault("backend.aerospike.hosts", []string{})
	v.SetDefault("backend.aerospike.port", 0)
	v.SetDefault("backend.aerospike.namespace", "")
	v.SetDefault("backend.aerospike.user", "")
	v.SetDefault("backend.aerospike.password", "")
	v.SetDefault("backend.aerospike.default_ttl_seconds", 0)
	v.SetDefault("backend.aerospike.max_read_retries", 2)
	v.SetDefault("backend.aerospike.max_write_retries", 0)
	v.SetDefault("backend.aerospike.connection_idle_timeout_seconds", 0)
	v.SetDefault("backend.aerospike.connection_queue_size", 0)
	v.SetDefault("backend.cassandra.hosts", "")
	v.SetDefault("backend.cassandra.keyspace", "")
	v.SetDefault("backend.cassandra.default_ttl_seconds", utils.CASSANDRA_DEFAULT_TTL_SECONDS)
	v.SetDefault("backend.memcache.hosts", []string{})
	v.SetDefault("backend.redis.host", "")
	v.SetDefault("backend.redis.port", 0)
	v.SetDefault("backend.redis.password", "")
	v.SetDefault("backend.redis.db", 0)
	v.SetDefault("backend.redis.expiration", utils.REDIS_DEFAULT_EXPIRATION_MINUTES)
	v.SetDefault("backend.redis.tls.enabled", false)
	v.SetDefault("backend.redis.tls.insecure_skip_verify", false)
	v.SetDefault("compression.type", "snappy")
	v.SetDefault("metrics.influx.enabled", false)
	v.SetDefault("metrics.influx.host", "")
	v.SetDefault("metrics.influx.database", "")
	v.SetDefault("metrics.influx.measurement", "")
	v.SetDefault("metrics.influx.username", "")
	v.SetDefault("metrics.influx.password", "")
	v.SetDefault("metrics.influx.align_timestamps", false)
	v.SetDefault("metrics.prometheus.port", 0)
	v.SetDefault("metrics.prometheus.namespace", "")
	v.SetDefault("metrics.prometheus.subsystem", "")
	v.SetDefault("metrics.prometheus.timeout_ms", 0)
	v.SetDefault("metrics.prometheus.enabled", false)
	v.SetDefault("rate_limiter.enabled", true)
	v.SetDefault("rate_limiter.num_requests", utils.RATE_LIMITER_NUM_REQUESTS)
	v.SetDefault("request_limits.allow_setting_keys", false)
	v.SetDefault("request_limits.max_size_bytes", utils.REQUEST_MAX_SIZE_BYTES)
	v.SetDefault("request_limits.max_num_values", utils.REQUEST_MAX_NUM_VALUES)
	v.SetDefault("request_limits.max_ttl_seconds", utils.REQUEST_MAX_TTL_SECONDS)
	v.SetDefault("routes.allow_public_write", true)
}

func setConfigFilePath(v *viper.Viper, filename string) {
	v.SetConfigName(filename)              // name of config file (without extension)
	v.AddConfigPath("/etc/prebid-cache/")  // path to look for the config file in
	v.AddConfigPath("$HOME/.prebid-cache") // call multiple times to add many search paths
	v.AddConfigPath(".")
}

func setEnvVarsLookup(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("PBC")
	v.AutomaticEnv()
}

type Configuration struct {
	Port           int               `mapstructure:"port"`
	AdminPort      int               `mapstructure:"admin_port"`
	IndexResponse  string            `mapstructure:"index_response"`
	Log            Log               `mapstructure:"log"`
	RateLimiting   RateLimiting      `mapstructure:"rate_limiter"`
	RequestLimits  RequestLimits     `mapstructure:"request_limits"`
	StatusResponse string            `mapstructure:"status_response"`
	Backend        Backend           `mapstructure:"backend"`
	Compression    Compression       `mapstructure:"compression"`
	Metrics        Metrics           `mapstructure:"metrics"`
	Routes         Routes            `mapstructure:"routes"`
	Stats          stats.StatsConfig `mapstructure:"stats"`
	Server         Server            `mapstructure:"server"`
	OWLog          OWLog             `mapstructure:"ow_log"`
}

// ValidateAndLog validates the config, terminating the program on any errors.
// It also logs the config values that it used.
func (cfg *Configuration) ValidateAndLog() {
	logger.Info("config.port: %d", cfg.Port)
	logger.Info("config.admin_port: %d", cfg.AdminPort)
	cfg.Log.validateAndLog()
	cfg.RateLimiting.validateAndLog()
	cfg.RequestLimits.validateAndLog()

	if err := cfg.Backend.validateAndLog(); err != nil {
		logger.Fatal("%s", err.Error())
	}

	cfg.Compression.validateAndLog()
	cfg.Metrics.validateAndLog()
	cfg.Server.validateAndLog()
	cfg.OWLog.validateAndLog()
	cfg.Routes.validateAndLog()
	validateAndLogStats(cfg.Stats)
}

type Log struct {
	Level LogLevel `mapstructure:"level"`
}

func (cfg *Log) validateAndLog() {
	logger.Info("config.log.level: %s", cfg.Level)
}

type LogLevel string

const (
	Trace   LogLevel = "trace"
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
	Fatal   LogLevel = "fatal"
	Panic   LogLevel = "panic"
)

type RateLimiting struct {
	Enabled              bool  `mapstructure:"enabled"`
	MaxRequestsPerSecond int64 `mapstructure:"num_requests"`
}

func (cfg *RateLimiting) validateAndLog() {
	logger.Info("config.rate_limiter.enabled: %t", cfg.Enabled)
	logger.Info("config.rate_limiter.num_requests: %d", cfg.MaxRequestsPerSecond)
}

type RequestLimits struct {
	MaxSize          int  `mapstructure:"max_size_bytes"`
	MaxNumValues     int  `mapstructure:"max_num_values"`
	MaxTTLSeconds    int  `mapstructure:"max_ttl_seconds"`
	AllowSettingKeys bool `mapstructure:"allow_setting_keys"`
}

func (cfg *RequestLimits) validateAndLog() {
	logger.Info("config.request_limits.allow_setting_keys: %v", cfg.AllowSettingKeys)

	if cfg.MaxTTLSeconds >= 0 {
		logger.Info("config.request_limits.max_ttl_seconds: %d", cfg.MaxTTLSeconds)
	} else {
		logger.Fatal("invalid config.request_limits.max_ttl_seconds: %d. Value cannot be negative.", cfg.MaxTTLSeconds)
	}

	if cfg.MaxSize >= 0 {
		logger.Info("config.request_limits.max_size_bytes: %d", cfg.MaxSize)
	} else {
		logger.Fatal("invalid config.request_limits.max_size_bytes: %d. Value cannot be negative.", cfg.MaxSize)
	}

	if cfg.MaxNumValues >= 0 {
		logger.Info("config.request_limits.max_num_values: %d", cfg.MaxNumValues)
	} else {
		logger.Fatal("invalid config.request_limits.max_num_values: %d. Value cannot be negative.", cfg.MaxNumValues)
	}
}

type Compression struct {
	Type CompressionType `mapstructure:"type"`
}

func (cfg *Compression) validateAndLog() {
	switch cfg.Type {
	case CompressionNone:
		fallthrough
	case CompressionSnappy:
		logger.Info("config.compression.type: %s", cfg.Type)
	default:
		logger.Fatal(`invalid config.compression.type: %s. It must be "none" or "snappy"`, cfg.Type)
	}
}

type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionSnappy CompressionType = "snappy"
)

type Metrics struct {
	Type       MetricsType       `mapstructure:"type"`
	Influx     InfluxMetrics     `mapstructure:"influx"`
	Prometheus PrometheusMetrics `mapstructure:"prometheus"`
}

func (cfg *Metrics) validateAndLog() {

	if cfg.Type == MetricsInflux || cfg.Influx.Enabled {
		cfg.Influx.validateAndLog()
		cfg.Influx.Enabled = true
	}

	if cfg.Prometheus.Enabled {
		cfg.Prometheus.validateAndLog()
		cfg.Prometheus.Enabled = true
	}

	metricsEnabled := cfg.Influx.Enabled || cfg.Prometheus.Enabled
	if cfg.Type == MetricsNone || cfg.Type == "" {
		if !metricsEnabled {
			logger.Info("Prebid Cache will run without metrics")
		}
	} else if cfg.Type != MetricsInflux {
		// Was any other metrics system besides "InfluxDB" or "Prometheus" specified in `cfg.Type`?
		if metricsEnabled {
			// Prometheus, Influx or both, are enabled. Log a message explaining that `prebid-cache` will
			// continue with supported metrics and non-supported metrics will be disabled
			logger.Info("Prebid Cache will run without unsupported metrics \"%s\".", cfg.Type)
		} else {
			// The only metrics engine specified in the configuration file is a non-supported
			// metrics engine. We should log error and exit program
			logger.Fatal("Metrics \"%s\" are not supported, exiting program.", cfg.Type)
		}
	}
}

type MetricsType string

const (
	MetricsNone   MetricsType = "none"
	MetricsInflux MetricsType = "influx"
)

type InfluxMetrics struct {
	Enabled         bool   `mapstructure:"enabled"`
	Host            string `mapstructure:"host"`
	Database        string `mapstructure:"database"`
	Measurement     string `mapstructure:"measurement"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	AlignTimestamps bool   `mapstructure:"align_timestamps"`
}

func (influxMetricsConfig *InfluxMetrics) validateAndLog() {
	// validate
	if influxMetricsConfig.Host == "" {
		logger.Fatal(`Despite being enabled, influx metrics came with no host info: config.metrics.influx.host = "".`)
	}
	if influxMetricsConfig.Database == "" {
		logger.Fatal(`Despite being enabled, influx metrics came with no database info: config.metrics.influx.database = "".`)
	}
	if influxMetricsConfig.Measurement == "" {
		log.Fatalf(`Despite being enabled, influx metrics came with no measurement info: config.metrics.influx.measurement = "".`)
	}

	// log
	logger.Info("config.metrics.influx.host: %s", influxMetricsConfig.Host)
	logger.Info("config.metrics.influx.database: %s", influxMetricsConfig.Database)
	logger.Info("config.metrics.influx.measurement: %s", influxMetricsConfig.Measurement)
	logger.Info("config.metrics.influx.align_timestamps: %v", influxMetricsConfig.AlignTimestamps)
}

type PrometheusMetrics struct {
	Port             int    `mapstructure:"port"`
	Namespace        string `mapstructure:"namespace"`
	Subsystem        string `mapstructure:"subsystem"`
	TimeoutMillisRaw int    `mapstructure:"timeout_ms"`
	Enabled          bool   `mapstructure:"enabled"`
}

// validateAndLog will error out when the value of port is 0
func (promMetricsConfig *PrometheusMetrics) validateAndLog() {
	if promMetricsConfig.Port == 0 {
		logger.Fatal(`Despite being enabled, prometheus metrics came with an empty port number: config.metrics.prometheus.port = 0`)
	}

	logger.Info("config.metrics.prometheus.namespace: %s", promMetricsConfig.Namespace)
	logger.Info("config.metrics.prometheus.subsystem: %s", promMetricsConfig.Subsystem)
	logger.Info("config.metrics.prometheus.port: %d", promMetricsConfig.Port)
}

func (promMetricsConfig *PrometheusMetrics) Timeout() time.Duration {
	return time.Duration(promMetricsConfig.TimeoutMillisRaw) * time.Millisecond
}

type Routes struct {
	AllowPublicWrite bool `mapstructure:"allow_public_write"`
}

func (cfg *Routes) validateAndLog() {
	if !cfg.AllowPublicWrite {
		logger.Info("Main server will only accept GET requests")
	}
}

func validateAndLogStats(stats stats.StatsConfig) {
	logger.Info("config.stats.host: %s", stats.Host)
	logger.Info("config.stats.port: %s", stats.Port)
	logger.Info("config.stats.dc_name: %s", stats.DCName)
	logger.Info("config.stats.default_hostname: %s", stats.DefaultHostName)
	logger.Info("config.stats.use_hostname: %v", stats.UseHostName)
	logger.Info("config.stats.publisher_interval: %d", stats.PublishInterval)
	logger.Info("config.stats.publisher_threshold: %d", stats.PublishThreshold)
	logger.Info("config.stats.retries: %d", stats.Retries)
	logger.Info("config.stats.dial_timeout: %d", stats.DialTimeout)
	logger.Info("config.stats.keep_alive_duration: %d", stats.KeepAliveDuration)
	logger.Info("config.stats.max_idle_connections: %d", stats.MaxIdleConnections)
	logger.Info("config.stats.max_idle_connections_per_host: %d", stats.MaxIdleConnectionsPerHost)
}

type Server struct {
	ServerPort string `mapstructure:"port"`
	ServerName string `mapstructure:"name"`
}

func (cfg *Server) validateAndLog() {
	logger.Info("config.server.port: %s", cfg.ServerPort)
	logger.Info("config.server.name: %s", cfg.ServerName)
}

type OWLog struct {
	LogLevel        logger.LogLevel `mapstructure:"level"`
	LogPath         string          `mapstructure:"path"`
	LogRotationTime time.Duration   `mapstructure:"rotation_time"`
	MaxLogFiles     int             `mapstructure:"max_log_files"`
	MaxLogSize      uint64          `mapstructure:"max_log_size"`
}

func (cfg *OWLog) validateAndLog() {
	logger.Info("config.ow_log.level: %v", cfg.LogLevel)
	logger.Info("config.ow_log.path: %s", cfg.LogPath)
	logger.Info("config.ow_log.rotation_time: %v", cfg.LogRotationTime)
	logger.Info("config.ow_log.max_log_files: %v", cfg.MaxLogFiles)
	logger.Info("config.ow_log.max_log_size: %v", cfg.MaxLogSize)
}
