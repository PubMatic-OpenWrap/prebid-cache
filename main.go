package main

import (
	_ "net/http/pprof"

	backendConfig "github.com/PubMatic-OpenWrap/prebid-cache/backends/config"
	"github.com/PubMatic-OpenWrap/prebid-cache/config"
	"github.com/PubMatic-OpenWrap/prebid-cache/endpoints/routing"
	"github.com/PubMatic-OpenWrap/prebid-cache/metrics"
	"github.com/PubMatic-OpenWrap/prebid-cache/server"
)

func main() {
	//log.SetOutput(os.Stdout)
	cfg := config.NewConfig()
	//setLogLevel(cfg.Log.Level)
	cfg.ValidateAndLog()

	appMetrics := metrics.CreateMetrics()
	backend := backendConfig.NewBackend(cfg, appMetrics)
	handler := routing.NewHandler(cfg, backend, appMetrics)
	go appMetrics.Export(cfg.Metrics)
	server.Listen(cfg, handler, appMetrics.Connections)
}

/*func setLogLevel(logLevel config.LogLevel) {
	level, err := log.ParseLevel(string(logLevel))
	if err != nil {
		log.Fatalf("Invalid logrus level: %v", err)
	}
	log.SetLevel(level)
}*/
