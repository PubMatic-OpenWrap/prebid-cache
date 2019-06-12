package stats

import (
	"fmt"

	"github.com/PubMatic-OpenWrap/prebid-cache/constant"

	"git.pubmatic.com/PubMatic/go-common.git/logger"
	"git.pubmatic.com/PubMatic/go-common.git/stats"
)

var S *stats.S

func InitStat(statIP, statPort, statServer, dc string) {
	statURL := statIP + ":" + statPort
	S = stats.NewStats(statURL, statServer, dc)
	if S == nil {
		logger.Error("Falied to Connect Stat Server ")
	}
}

func LogCacheFailedGetStats(errorString string) {
	fmt.Printf(constant.StatsKeyCacheFailedGet, errorString)
	S.Increment(fmt.Sprintf(constant.StatsKeyCacheFailedGet, errorString),
		constant.StatsKeyCacheFailedGetCutoff, 1)
}

func LogCacheMissStats() {
	S.Increment(fmt.Sprintf(constant.StatsKeyCacheMiss),
		constant.StatsKeyCacheMissCutOff, 1)
}

func LogCacheFailedPutStats(errorString string) {
	S.Increment(fmt.Sprintf(constant.StatsKeyCacheFailedPut, errorString),
		constant.StatsKeyCacheFailedPutCutoff, 1)
}

func LogCacheRequestedGetStats() {
	S.Increment(fmt.Sprintf(constant.StatsKeyCacheRequestedGet),
		constant.StatsKeyCacheRequestedGetCutoff, 1)
}

func LogCacheRequestedPutStats() {
	S.Increment(fmt.Sprintf(constant.StatsKeyCacheRequestedPut),
		constant.StatsKeyCacheRequestedPutCutoff, 1)
}

func LogAerospikeErrorStats() {
	S.Increment(fmt.Sprintf(constant.StatsKeyAerospikeCreationError),
		constant.StatsKeyAerospikeCreationErrorCutoff, 1)

}
