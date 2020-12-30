package stats

import (
	"fmt"
	"git.pubmatic.com/PubMatic/go-common.git/logger"
	stats "git.pubmatic.com/PubMatic/go-common.git/stats2"
	"github.com/PubMatic-OpenWrap/prebid-cache/constant"
)

var sc *stats.Client

type statLogger struct{}

func (l statLogger) Error(format string, args ...interface{}) {
	logger.Error(format, args...)
}

func (l statLogger) Info(format string, args ...interface{}) {
	logger.Info(format, args...)
}

func InitStat(statIP, statPort, statServer, dc string) {

	cgf := stats.Config{
		Host:   statIP,
		Port:   statPort,
		Server: statServer,
		DC:     dc,
	}

	var err error
	sc, err = stats.NewClient(cgf, statLogger{})
	if err != nil {
		logger.Error("failed to initialize stats client")
	}
}

func LogCacheFailedGetStats(errorString string) {
	fmt.Printf(constant.StatsKeyCacheFailedGet, errorString)
	sc.PublishStat(constant.StatsKeyCacheFailedGet, 1, errorString)
}

func LogCacheMissStats() {
	sc.PublishStat(constant.StatsKeyCacheMiss, 1)
}

func LogCacheFailedPutStats(errorString string) {
	sc.PublishStat(constant.StatsKeyCacheFailedPut, 1, errorString)
}

func LogCacheRequestedGetStats() {
	sc.PublishStat(constant.StatsKeyCacheRequestedGet, 1)
}

func LogCacheRequestedPutStats() {
	sc.PublishStat(constant.StatsKeyCacheRequestedPut, 1)
}

func LogAerospikeErrorStats() {
	sc.PublishStat(constant.StatsKeyAerospikeCreationError, 1)
}
