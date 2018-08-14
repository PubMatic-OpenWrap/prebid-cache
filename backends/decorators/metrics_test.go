package decorators

import (
	"context"
	"fmt"
	"testing"

	"github.com/prebid/prebid-cache/backends"
	"github.com/prebid/prebid-cache/metrics"
	"github.com/prebid/prebid-cache/metrics/metricstest"
	uuid "github.com/satori/go.uuid"
)

type failedBackend struct{}

func (b *failedBackend) Get(ctx context.Context, key string, rqID string) (string, error) {
	return "", fmt.Errorf("Failure")
}

func (b *failedBackend) Put(ctx context.Context, key string, value string, rqID string) error {
	return fmt.Errorf("Failure")
}

func TestGetSuccessMetrics(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	rawBackend := backends.NewMemoryBackend()
	rawBackend.Put(context.Background(), "foo", "xml<vast></vast>", rqID)
	backend := LogMetrics(rawBackend, m)
	backend.Get(context.Background(), "foo", rqID)

	metricstest.AssertSuccessMetricsExist(t, m.GetsBackend)
}

func TestGetErrorMetrics(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	backend := LogMetrics(&failedBackend{}, m)
	backend.Get(context.Background(), "foo", rqID)

	metricstest.AssertErrorMetricsExist(t, m.GetsBackend)
}

func TestPutSuccessMetrics(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "xml<vast></vast>", rqID)

	assertSuccessMetricsExist(t, m.PutsBackend)
	if m.PutsBackend.XmlRequest.Count() != 1 {
		t.Errorf("An xml request should have been logged.")
	}
}

func TestPutErrorMetrics(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	backend := LogMetrics(&failedBackend{}, m)
	backend.Put(context.Background(), "foo", "xml<vast></vast>", rqID)

	assertErrorMetricsExist(t, m.PutsBackend)
	if m.PutsBackend.XmlRequest.Count() != 1 {
		t.Errorf("The request should have been counted.")
	}
}

func TestJsonPayloadMetrics(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "json{\"key\":\"value\"", rqID)
	backend.Get(context.Background(), "foo", rqID)

	if m.PutsBackend.JsonRequest.Count() != 1 {
		t.Errorf("A json Put should have been logged.")
	}
}

func TestPutSizeSampling(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	payload := `json{"key":"value"}`
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", payload, rqID)

	if m.PutsBackend.RequestLength.Count() != 1 {
		t.Errorf("A request size sample should have been logged.")
	}
}

func TestInvalidPayloadMetrics(t *testing.T) {
	rqID := uuid.NewV4().String()
	m := metrics.CreateMetrics()
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "bar", rqID)
	backend.Get(context.Background(), "foo", rqID)

	if m.PutsBackend.InvalidRequest.Count() != 1 {
		t.Errorf("A Put request of invalid format should have been logged.")
	}
}

func assertSuccessMetricsExist(t *testing.T, entry *metrics.MetricsEntryByFormat) {
	t.Helper()
	if entry.Duration.Count() != 1 {
		t.Errorf("The request duration should have been counted.")
	}
	if entry.BadRequest.Count() != 0 {
		t.Errorf("No Bad requests should have been counted.")
	}
	if entry.Errors.Count() != 0 {
		t.Errorf("No Errors should have been counted.")
	}
}

func assertErrorMetricsExist(t *testing.T, entry *metrics.MetricsEntryByFormat) {
	t.Helper()
	if entry.Duration.Count() != 0 {
		t.Errorf("The request duration should not have been counted.")
	}
	if entry.BadRequest.Count() != 0 {
		t.Errorf("No Bad requests should have been counted.")
	}
	if entry.Errors.Count() != 1 {
		t.Errorf("An Error should have been counted.")
	}
}
