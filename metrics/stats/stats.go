package stats

import (
	"github.com/prebid/prebid-cache/utils"
)

var (
	owStats statsMetrics
)

func InitStat(statsCfg *StatsConfig) {

	serverName := statsCfg.DefaultHostName
	if statsCfg.UseHostName {
		serverName = utils.GetServerName()
	}

	var err error
	if owStats, err = newStatsClient(statsCfg, serverName); err != nil {
		owStats = noStats{}
	}
}

type statsMetrics interface {
	LogCacheFailedGetStats(errorString string)
	LogCacheMissStats()
	LogCacheFailedPutStats(errorString string)
	LogCacheRequestedGetStats()
	LogCacheRequestedPutStats()
	LogAerospikeErrorStats()
}

func LogCacheFailedGetStats(errorString string) {
	owStats.LogCacheFailedGetStats(errorString)
}

func LogCacheMissStats() {
	owStats.LogCacheMissStats()
}

func LogCacheFailedPutStats(errorString string) {
	owStats.LogCacheFailedPutStats(errorString)
}

func LogCacheRequestedGetStats() {
	owStats.LogCacheRequestedGetStats()
}

func LogCacheRequestedPutStats() {
	owStats.LogCacheRequestedPutStats()
}

func LogAerospikeErrorStats() {
	owStats.LogAerospikeErrorStats()
}
