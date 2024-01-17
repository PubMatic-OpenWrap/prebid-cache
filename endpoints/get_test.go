package endpoints

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/prebid-cache/backends"
	"github.com/prebid/prebid-cache/metrics"
	"github.com/prebid/prebid-cache/metrics/metricstest"
	"github.com/prebid/prebid-cache/metrics/stats"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func init() {
	stats.InitStat(&stats.StatsConfig{})
}

func TestGetInvalidUUIDs(t *testing.T) {
	backend := backends.NewMemoryBackend()
	router := httprouter.New()

	mockMetrics := metricstest.CreateMockMetrics()
	m := &metrics.Metrics{
		MetricEngines: []metrics.CacheMetrics{
			&mockMetrics,
		},
	}

	router.GET("/cache", NewGetHandler(backend, m, false))

	getResults := doMockGet(t, router, "fdd9405b-ef2b-46da-a55a-2f526d338e16")
	if getResults.Code != http.StatusNotFound {
		t.Fatalf("Expected GET to return 404 on unrecognized ID. Got: %d", getResults.Code)
		return
	}

	getResults = doMockGet(t, router, "abc")
	if getResults.Code != http.StatusNotFound {
		t.Fatalf("Expected GET to return 404 on unrecognized ID. Got: %d", getResults.Code)
		return
	}
}

func TestGetHandler(t *testing.T) {
	preExistentDataInBackend := map[string]string{
		"non-36-char-key-maps-to-json":         `json{"field":"value"}`,
		"36-char-key-maps-to-non-xml-nor-json": `#@!*{"desc":"data got malformed and is not prefixed with 'xml' nor 'json' substring"}`,
		"36-char-key-maps-to-actual-xml-value": "xml<tag>xml data here</tag>",
	}

	type logEntry struct {
		msg string
		lvl logrus.Level
	}
	type testInput struct {
		uuid      string
		allowKeys bool
	}
	type testOutput struct {
		responseCode    int
		responseBody    string
		logEntries      []logEntry
		expectedMetrics []string
	}

	testCases := []struct {
		desc string
		in   testInput
		out  testOutput
	}{
		{
			"Configuration that allows custom keys. These are not required to be 36 char long. Since the uuid maps to a value, return it along a 200 status code",
			testInput{
				uuid:      "non-36-char-key-maps-to-json",
				allowKeys: true,
			},
			testOutput{
				responseCode: http.StatusOK,
				responseBody: `{"field":"value"}`,
				logEntries:   []logEntry{},
				expectedMetrics: []string{
					"RecordGetTotal",
					"RecordGetDuration",
				},
			},
		},
		{
			"Valid 36 char long UUID returns valid XML. Don't return nor log error",
			testInput{uuid: "36-char-key-maps-to-actual-xml-value"},
			testOutput{
				responseCode: http.StatusOK,
				responseBody: "<tag>xml data here</tag>",
				logEntries:   []logEntry{},
				expectedMetrics: []string{
					"RecordGetTotal",
					"RecordGetDuration",
				},
			},
		},
	}

	// Lower Log Treshold so we can see DebugLevel entries in our mock logrus log
	logrus.SetLevel(logrus.DebugLevel)

	// Test suite-wide objects
	hook := test.NewGlobal()

	defer func() { logrus.StandardLogger().ExitFunc = nil }()
	var fatal bool
	logrus.StandardLogger().ExitFunc = func(int) { fatal = true }

	for _, test := range testCases {
		// Reset the fatal flag to false every test
		fatal = false

		// Set up test object
		backend, err := backends.NewMemoryBackendWithValues(preExistentDataInBackend)
		if !assert.NoError(t, err, "%s. Mock backend could not be created", test.desc) {
			hook.Reset()
			continue
		}
		router := httprouter.New()
		mockMetrics := metricstest.CreateMockMetrics()
		m := &metrics.Metrics{
			MetricEngines: []metrics.CacheMetrics{
				&mockMetrics,
			},
		}
		router.GET("/cache", NewGetHandler(backend, m, test.in.allowKeys))

		// Run test
		getResults := httptest.NewRecorder()

		body := new(bytes.Buffer)
		getReq, err := http.NewRequest("GET", "/cache"+"?uuid="+test.in.uuid, body)
		if !assert.NoError(t, err, "Failed to create a GET request: %v", err) {
			hook.Reset()
			continue
		}
		router.ServeHTTP(getResults, getReq)

		// Assert server response and status code
		assert.Equal(t, test.out.responseCode, getResults.Code, test.desc)
		assert.Equal(t, test.out.responseBody, getResults.Body.String(), test.desc)

		// Assert log entries
		if assert.Len(t, hook.Entries, len(test.out.logEntries), test.desc) {
			for i := 0; i < len(test.out.logEntries); i++ {
				assert.Equal(t, test.out.logEntries[i].msg, hook.Entries[i].Message, test.desc)
				assert.Equal(t, test.out.logEntries[i].lvl, hook.Entries[i].Level, test.desc)
			}
			// Assert the logger didn't exit the program
			assert.False(t, fatal, test.desc)
		}

		// Assert recorded metrics
		metricstest.AssertMetrics(t, test.out.expectedMetrics, mockMetrics)

		// Reset log
		hook.Reset()
	}
}
