package stats

type StatsConfig struct {
	Host                      string `mapstructure:"host"`
	Port                      string `mapstructure:"port"`
	DCName                    string `mapstructure:"dc_name"`
	DefaultHostName           string `mapstructure:"default_hostname"`
	UseHostName               bool   `mapstructure:"use_hostname"`
	PublishInterval           int    `mapstructure:"publish_interval"`
	PublishThreshold          int    `mapstructure:"publish_threshold"`
	Retries                   int    `mapstructure:"retries"`
	DialTimeout               int    `mapstructure:"dial_timeout"`
	KeepAliveDuration         int    `mapstructure:"keep_alive_duration"`
	MaxIdleConnections        int    `mapstructure:"max_idle_connections"`
	MaxIdleConnectionsPerHost int    `mapstructure:"max_idle_connections_per_host"`
}
