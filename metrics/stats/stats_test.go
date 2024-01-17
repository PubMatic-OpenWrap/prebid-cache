package stats

import (
	"testing"
)

func TestStatsInit(t *testing.T) {
	InitStat(&StatsConfig{
		Host:                      "127.0.0.1",
		Port:                      "8080",
		DCName:                    "TestDC",
		DefaultHostName:           "N:P",
		UseHostName:               false,
		PublishInterval:           2,
		PublishThreshold:          20000,
		Retries:                   3,
		DialTimeout:               5,
		KeepAliveDuration:         2,
		MaxIdleConnections:        2,
		MaxIdleConnectionsPerHost: 2,
	})
}

func TestStatsLogCacheFailedGetStats(t *testing.T) {
	LogCacheFailedGetStats("Error string")
}

func TestStatsLogCacheFailedPutStats(t *testing.T) {
	LogCacheFailedPutStats("Error string")
}

func TestStatsLogCacheRequestedGetStats(t *testing.T) {
	LogCacheRequestedGetStats()
}

func TestStatsLogCacheMissStats(t *testing.T) {
	LogCacheMissStats()
}

func TestStatsLogCacheRequestedPutStats(t *testing.T) {
	LogCacheRequestedPutStats()
}

func TestStatsLogAerospikeErrorStats(t *testing.T) {
	LogAerospikeErrorStats()
}
