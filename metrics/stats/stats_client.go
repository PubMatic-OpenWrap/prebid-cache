package stats

import (
	"fmt"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"git.pubmatic.com/PubMatic/go-common/tcpstats"
	"github.com/prebid/prebid-cache/constant"
)

type statLogger struct{}

func (l statLogger) Error(format string, args ...interface{}) {
	logger.Error(format, args...)
}

func (l statLogger) Info(format string, args ...interface{}) {
	logger.Info(format, args...)
}

type statsClient struct {
	client *tcpstats.Client
}

func newStatsClient(config *StatsConfig, server string) (*statsClient, error) {

	cgf := tcpstats.Config{
		Host:                config.Host,
		Port:                config.Port,
		Server:              server,
		DC:                  config.DCName,
		PublishingInterval:  config.PublishInterval,
		PublishingThreshold: config.PublishThreshold,
		Retries:             config.Retries,
		DialTimeout:         config.DialTimeout,
		KeepAliveDuration:   config.KeepAliveDuration,
		MaxIdleConns:        config.MaxIdleConnections,
		MaxIdleConnsPerHost: config.MaxIdleConnectionsPerHost,
	}

	sc, err := tcpstats.NewClient(cgf, statLogger{})
	if err != nil {
		return nil, err
	}

	return &statsClient{client: sc}, nil
}

func (st *statsClient) LogCacheFailedGetStats(errorString string) {
	st.client.PublishStat(fmt.Sprintf(constant.StatsKeyCacheFailedGet, errorString), 1)
}

func (st *statsClient) LogCacheMissStats() {
	st.client.PublishStat(constant.StatsKeyCacheMiss, 1)
}

func (st *statsClient) LogCacheFailedPutStats(errorString string) {
	st.client.PublishStat(fmt.Sprintf(constant.StatsKeyCacheFailedPut, errorString), 1)
}

func (st *statsClient) LogCacheRequestedGetStats() {
	st.client.PublishStat(constant.StatsKeyCacheRequestedGet, 1)
}

func (st *statsClient) LogCacheRequestedPutStats() {
	st.client.PublishStat(constant.StatsKeyCacheRequestedPut, 1)
}

func (st *statsClient) LogAerospikeErrorStats() {
	st.client.PublishStat(constant.StatsKeyAerospikeCreationError, 1)
}
