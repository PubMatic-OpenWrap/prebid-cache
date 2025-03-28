package metrics

import (
	"fmt"
	"time"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/prebid/prebid-cache/config"
	"github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

const TenSeconds time.Duration = time.Second * 10

const MetricsInfluxDB = "InfluxDB"

type InfluxMetrics struct {
	Registry    metrics.Registry
	Puts        *InfluxMetricsEntry
	Gets        *InfluxMetricsEntry
	PutsBackend *InfluxMetricsEntryByFormat
	GetsBackend *InfluxMetricsEntry
	GetsErr     *InfluxMetricsGetErrors
	Connections *InfluxConnectionMetrics
	MetricsName string
}

type InfluxMetricsEntry struct {
	Duration   metrics.Timer
	Errors     metrics.Meter
	BadRequest metrics.Meter
	Request    metrics.Meter
	Update     metrics.Meter
}

type InfluxMetricsEntryByFormat struct {
	Duration       metrics.Timer
	Request        metrics.Meter
	Errors         metrics.Meter
	BadRequest     metrics.Meter
	JsonRequest    metrics.Meter
	XmlRequest     metrics.Meter
	InvalidRequest metrics.Meter
	RequestLength  metrics.Histogram
	RequestTTL     metrics.Timer
}

type InfluxConnectionMetrics struct {
	ActiveConnections      metrics.Counter
	ConnectionCloseErrors  metrics.Meter
	ConnectionAcceptErrors metrics.Meter
}

type InfluxMetricsGetErrors struct {
	KeyNotFoundErrors metrics.Meter
	MissingKeyErrors  metrics.Meter
}

func NewInfluxGetErrorMetrics(name string, r metrics.Registry) *InfluxMetricsGetErrors {
	return &InfluxMetricsGetErrors{
		KeyNotFoundErrors: metrics.GetOrRegisterMeter(fmt.Sprintf("%s.key_not_found", name), r),
		MissingKeyErrors:  metrics.GetOrRegisterMeter(fmt.Sprintf("%s.missing_key", name), r),
	}
}

// NewInfluxMetricsEntryGet initializes all the metrics of InfluxMetricsEntry except for
// Update which makes no sense in the context of a Get call
func NewInfluxMetricsEntryGet(name string, r metrics.Registry) *InfluxMetricsEntry {
	return &InfluxMetricsEntry{
		Duration:   metrics.GetOrRegisterTimer(fmt.Sprintf("%s.request_duration", name), r),
		Errors:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.error_count", name), r),
		BadRequest: metrics.GetOrRegisterMeter(fmt.Sprintf("%s.bad_request_count", name), r),
		Request:    metrics.GetOrRegisterMeter(fmt.Sprintf("%s.request_count", name), r),
	}
}

// NewInfluxMetricsEntryEndpointPuts initializes all the metrics of InfluxMetricsEntry including
// Update which will account for the Put requests that come with their own Key to store the value in.
func NewInfluxMetricsEntryEndpointPuts(name string, r metrics.Registry) *InfluxMetricsEntry {
	return &InfluxMetricsEntry{
		Duration:   metrics.GetOrRegisterTimer(fmt.Sprintf("%s.request_duration", name), r),
		Errors:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.error_count", name), r),
		BadRequest: metrics.GetOrRegisterMeter(fmt.Sprintf("%s.bad_request_count", name), r),
		Request:    metrics.GetOrRegisterMeter(fmt.Sprintf("%s.request_count", name), r),
		Update:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.updated_key_count", name), r),
	}
}

func NewInfluxMetricsEntryBackendPuts(name string, r metrics.Registry) *InfluxMetricsEntryByFormat {
	return &InfluxMetricsEntryByFormat{
		Duration:       metrics.GetOrRegisterTimer(fmt.Sprintf("%s.request_duration", name), r),
		Errors:         metrics.GetOrRegisterMeter(fmt.Sprintf("%s.error_count", name), r),
		BadRequest:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.bad_request_count", name), r),
		JsonRequest:    metrics.GetOrRegisterMeter(fmt.Sprintf("%s.json_request_count", name), r),
		XmlRequest:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.xml_request_count", name), r),
		InvalidRequest: metrics.GetOrRegisterMeter(fmt.Sprintf("%s.unknown_request_count", name), r),
		RequestLength:  metrics.GetOrRegisterHistogram(name+".request_size_bytes", r, metrics.NewExpDecaySample(1028, 0.015)),
		RequestTTL:     metrics.GetOrRegisterTimer(fmt.Sprintf("%s.request_ttl_seconds", name), r),
	}
}

func NewInfluxConnectionMetrics(r metrics.Registry) *InfluxConnectionMetrics {
	return &InfluxConnectionMetrics{
		ActiveConnections:      metrics.GetOrRegisterCounter("connections.active_incoming", r),
		ConnectionAcceptErrors: metrics.GetOrRegisterMeter("connections.accept_errors", r),
		ConnectionCloseErrors:  metrics.GetOrRegisterMeter("connections.close_errors", r),
	}
}

func CreateInfluxMetrics() *InfluxMetrics {
	flushTime := TenSeconds
	r := metrics.NewPrefixedRegistry("prebidcache.")
	m := &InfluxMetrics{
		Registry:    r,
		Puts:        NewInfluxMetricsEntryEndpointPuts("puts.current_url", r),
		Gets:        NewInfluxMetricsEntryGet("gets.current_url", r),
		PutsBackend: NewInfluxMetricsEntryBackendPuts("puts.backend", r),
		GetsBackend: NewInfluxMetricsEntryGet("gets.backend", r),
		GetsErr:     NewInfluxGetErrorMetrics("gets.backend_error", r),
		Connections: NewInfluxConnectionMetrics(r),
		MetricsName: MetricsInfluxDB,
	}

	metrics.RegisterDebugGCStats(m.Registry)
	metrics.RegisterRuntimeMemStats(m.Registry)

	go metrics.CaptureRuntimeMemStats(m.Registry, flushTime)
	go metrics.CaptureDebugGCStats(m.Registry, flushTime)

	return m
}

// Export begins metric publishing services.
func (m InfluxMetrics) Export(cfg config.Metrics) {

	logger.Info("Metrics will be exported to Influx with host=%s, db=%s, username=%s", cfg.Influx.Host, cfg.Influx.Database, cfg.Influx.Username)
	influxdb.InfluxDB(
		m.Registry,
		TenSeconds,
		cfg.Influx.Host,
		cfg.Influx.Database,
		cfg.Influx.Measurement,
		cfg.Influx.Username,
		cfg.Influx.Password,
		cfg.Influx.AlignTimestamps,
	)
}

func (m *InfluxMetrics) GetEngineRegistry() interface{} {
	return &m.Registry
}

func (m *InfluxMetrics) GetMetricsEngineName() string {
	return m.MetricsName
}

func (m *InfluxMetrics) RecordPutError() {
	m.Puts.Errors.Mark(1)
}

func (m *InfluxMetrics) RecordPutBadRequest() {
	m.Puts.BadRequest.Mark(1)
}

func (m *InfluxMetrics) RecordPutTotal() {
	m.Puts.Request.Mark(1)
}

func (m *InfluxMetrics) RecordPutDuration(duration time.Duration) {
	m.Puts.Duration.Update(duration)
}

func (m *InfluxMetrics) RecordPutKeyProvided() {
	m.Puts.Update.Mark(1)
}

func (m *InfluxMetrics) RecordGetError() {
	m.Gets.Errors.Mark(1)
}

func (m *InfluxMetrics) RecordGetBadRequest() {
	m.Gets.BadRequest.Mark(1)
}

func (m *InfluxMetrics) RecordGetTotal() {
	m.Gets.Request.Mark(1)
}

func (m *InfluxMetrics) RecordGetDuration(duration time.Duration) {
	m.Gets.Duration.Update(duration)
}

func (m *InfluxMetrics) RecordPutBackendXml() {
	m.PutsBackend.XmlRequest.Mark(1)
}

func (m *InfluxMetrics) RecordPutBackendJson() {
	m.PutsBackend.JsonRequest.Mark(1)
}

func (m *InfluxMetrics) RecordPutBackendInvalid() {
	m.PutsBackend.InvalidRequest.Mark(1)
}

func (m *InfluxMetrics) RecordPutBackendDuration(duration time.Duration) {
	m.PutsBackend.Duration.Update(duration)
}

func (m *InfluxMetrics) RecordPutBackendTTLSeconds(duration time.Duration) {
	m.PutsBackend.RequestTTL.Update(duration)
}

func (m *InfluxMetrics) RecordPutBackendError() {
	m.PutsBackend.Errors.Mark(1)
}

func (m *InfluxMetrics) RecordGetBackendTotal() {
	m.GetsBackend.Request.Mark(1)
}

func (m *InfluxMetrics) RecordPutBackendSize(sizeInBytes float64) {
	m.PutsBackend.RequestLength.Update(int64(sizeInBytes))
}

func (m *InfluxMetrics) RecordGetBackendDuration(duration time.Duration) {
	m.GetsBackend.Duration.Update(duration)
}

func (m *InfluxMetrics) RecordGetBackendError() {
	m.GetsBackend.Errors.Mark(1)
}

func (m *InfluxMetrics) RecordKeyNotFoundError() {
	m.GetsErr.KeyNotFoundErrors.Mark(1)
}

func (m *InfluxMetrics) RecordMissingKeyError() {
	m.GetsErr.MissingKeyErrors.Mark(1)
}

func (m *InfluxMetrics) RecordConnectionOpen() {
	m.Connections.ActiveConnections.Inc(1)
}

func (m *InfluxMetrics) RecordConnectionClosed() {
	m.Connections.ActiveConnections.Dec(1)
}

func (m *InfluxMetrics) RecordCloseConnectionErrors() {
	m.Connections.ConnectionCloseErrors.Mark(1)
}

func (m *InfluxMetrics) RecordAcceptConnectionErrors() {
	m.Connections.ConnectionAcceptErrors.Mark(1)
}
