package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/PubMatic-OpenWrap/prebid-cache/backends"
	backendDecorators "github.com/PubMatic-OpenWrap/prebid-cache/backends/decorators"
	endpointDecorators "github.com/PubMatic-OpenWrap/prebid-cache/endpoints/decorators"
	"github.com/PubMatic-OpenWrap/prebid-cache/metrics/metricstest"
	"github.com/PubMatic-OpenWrap/prebid-cache/stats"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	stats.InitStat("127.0.0.1", "8888", "TestHost", "TestDC",
		"8080", 2, 20000, 3, 2, 5, 2, 2, true)
}

func TestJSONString(t *testing.T) {
	expectStored(
		t,
		"{\"puts\":[{\"type\":\"json\",\"value\":\"plain text\"}]}",
		"\"plain text\"",
		"application/json")
}

func TestEscapedString(t *testing.T) {
	expectStored(
		t,
		"{\"puts\":[{\"type\":\"json\",\"value\":\"esca\\\"ped\"}]}",
		"\"esca\\\"ped\"",
		"application/json")
}

func TestUnescapedString(t *testing.T) {
	expectFailedPut(t, "{\"puts\":[{\"type\":\"json\",\"value\":\"badly-esca\"ped\"}]}")
}

func TestNumber(t *testing.T) {
	expectStored(t, "{\"puts\":[{\"type\":\"json\",\"value\":5}]}", "5", "application/json")
}

func TestObject(t *testing.T) {
	expectStored(
		t,
		"{\"puts\":[{\"type\":\"json\",\"value\":{\"custom_key\":\"foo\"}}]}",
		"{\"custom_key\":\"foo\"}",
		"application/json")
}

func TestNull(t *testing.T) {
	expectStored(t, "{\"puts\":[{\"type\":\"json\",\"value\":null}]}", "null", "application/json")
}

func TestBoolean(t *testing.T) {
	expectStored(t, "{\"puts\":[{\"type\":\"json\",\"value\":true}]}", "true", "application/json")
}

func TestExtraProperty(t *testing.T) {
	expectStored(
		t,
		"{\"puts\":[{\"type\":\"json\",\"value\":null,\"irrelevant\":\"foo\"}]}",
		"null",
		"application/json")
}

func TestInvalidJSON(t *testing.T) {
	expectFailedPut(t, "{\"puts\":[{\"type\":\"json\",\"value\":malformed}]}")
}

func TestMissingProperty(t *testing.T) {
	expectFailedPut(t, "{\"puts\":[{\"type\":\"json\",\"unrecognized\":true}]}")
}

func TestMixedValidityPuts(t *testing.T) {
	expectFailedPut(t, "{\"puts\":[{\"type\":\"json\",\"value\":true}, {\"type\":\"json\",\"unrecognized\":true}]}")
}

func TestXMLString(t *testing.T) {
	expectStored(t, "{\"puts\":[{\"type\":\"xml\",\"value\":\"<tag></tag>\"}]}", "<tag></tag>", "application/xml")
}

func TestNonJsonNorXMLString(t *testing.T) {
	expectFailedPut(t, "{\"puts\":[{\"type\":\"yaml\",\"value\":\"<tag></tag>\"}]}")
}

func TestCrossScriptEscaping(t *testing.T) {
	expectStored(t, "{\"puts\":[{\"type\":\"xml\",\"value\":\"<tag>esc\\\"aped</tag>\"}]}", "<tag>esc\"aped</tag>", "application/xml")
}

func TestXMLOther(t *testing.T) {
	expectFailedPut(t, "{\"puts\":[{\"type\":\"xml\",\"value\":5}]}")
}

func TestGetInvalidUUIDs(t *testing.T) {
	backend := backends.NewMemoryBackend()
	router := httprouter.New()
	router.GET("/cache", NewGetHandler(backend, false))

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

func TestReadinessCheck(t *testing.T) {
	requestRecorder := httptest.NewRecorder()

	router := httprouter.New()
	router.GET("/status", Status)
	req, _ := http.NewRequest("GET", "/status", new(bytes.Buffer))
	router.ServeHTTP(requestRecorder, req)

	if requestRecorder.Code != http.StatusNoContent {
		t.Errorf("/status endpoint should always return a 204. Got %d", requestRecorder.Code)
	}
}

func TestNegativeTTL(t *testing.T) {
	// Input
	inReqBody := fmt.Sprintf("{\"puts\":[{\"type\":\"json\",\"value\":\"<tag>YourXMLcontentgoeshere.</tag>\",\"ttlseconds\":-1}]}")
	inRequest, err := http.NewRequest("POST", "/cache", strings.NewReader(inReqBody))
	assert.NoError(t, err, "Failed to create a POST request: %v", err)

	// Expected Values
	expectedErrorMsg := "Error request ttl seconds value must not be negative.\n"
	expectedStatusCode := http.StatusBadRequest

	// Set up server to run our test
	testRouter := httprouter.New()
	testBackend := backends.NewMemoryBackend()

	testRouter.POST("/cache", NewPutHandler(testBackend, 10, true))

	recorder := httptest.NewRecorder()

	// Run test
	testRouter.ServeHTTP(recorder, inRequest)

	// Assertions
	assert.Equal(t, expectedErrorMsg, recorder.Body.String(), "Put should have failed because we passed a negative ttlseconds value.\n")
	assert.Equalf(t, expectedStatusCode, recorder.Code, "Expected 400 response. Got: %d", recorder.Code)
}

func TestCustomKey(t *testing.T) {
	type aTest struct {
		desc         string
		inCustomKey  string
		expectedUuid string
	}
	testGroups := []struct {
		allowSettingKeys bool
		testCases        []aTest
	}{
		{
			allowSettingKeys: false,
			testCases: []aTest{
				{
					desc:         "Custom key maps to element in cache but setting keys is not allowed, set value with random UUID",
					inCustomKey:  "36-char-key-maps-to-actual-xml-value",
					expectedUuid: `[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`,
				},
				{
					desc:         "Custom key maps to no element in cache, set value with random UUID and respond 200",
					inCustomKey:  "36-char-key-maps-to-actual-xml-value",
					expectedUuid: `[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`,
				},
			},
		},
		{
			allowSettingKeys: true,
			testCases: []aTest{
				{
					desc:         "Setting keys allowed but key already maps to an element in cache, don't set value and respond with blank UUID",
					inCustomKey:  "36-char-key-maps-to-actual-xml-value",
					expectedUuid: "",
				},
				{
					desc:         "Custom key maps to no element in cache, set value and respond with 200 and the custom UUID",
					inCustomKey:  "cust-key-maps-to-no-value-in-backend",
					expectedUuid: "cust-key-maps-to-no-value-in-backend",
				},
			},
		},
	}

	mockBackendWithValues := newMockBackend()
	m := metricstest.CreateMockMetrics()

	for _, tgroup := range testGroups {
		for _, tc := range tgroup.testCases {
			// Instantiate prebid cache prod server with mock metrics and a mock metrics that
			// already contains some values
			router := httprouter.New()
			putEndpointHandler := NewPutHandler(mockBackendWithValues, 10, tgroup.allowSettingKeys)
			monitoredHandler := endpointDecorators.MonitorHttp(putEndpointHandler, m, endpointDecorators.PostMethod)
			router.POST("/cache", monitoredHandler)

			recorder := httptest.NewRecorder()

			reqBody := fmt.Sprintf(`{"puts":[{"type":"json","value":"xml<tag>updated_value</tag>","key":"%s"}]}`, tc.inCustomKey)
			request, err := http.NewRequest("POST", "/cache", strings.NewReader(reqBody))
			assert.NoError(t, err, "Test request could not be created")

			// Run test
			router.ServeHTTP(recorder, request)

			// Assert status code. All scenarios should return a 200 code
			assert.Equal(t, http.StatusOK, recorder.Code, tc.desc)

			// Assert response UUID
			if tc.expectedUuid == "" {
				assert.Equalf(t, `{"responses":[{"uuid":""}]}`, recorder.Body.String(), tc.desc)
			} else {
				re, err := regexp.Compile(tc.expectedUuid)
				assert.NoError(t, err, tc.desc)
				assert.Greater(t, len(re.Find(recorder.Body.Bytes())), 0, tc.desc)
			}
		}
	}
}

func TestRequestReadError(t *testing.T) {
	// Setup server and mock body request reader
	mockBackendWithValues := newMockBackend()
	putEndpointHandler := NewPutHandler(mockBackendWithValues, 10, false)

	router := httprouter.New()
	router.POST("/cache", putEndpointHandler)

	recorder := httptest.NewRecorder()

	// make our request body reader's Read() and Close() methods to return errors
	mockRequestReader := faultyRequestBodyReader{}
	mockRequestReader.On("Read", mock.AnythingOfType("[]uint8")).Return(0, errors.New("Read error"))
	mockRequestReader.On("Close").Return(errors.New("Read error"))

	request, _ := http.NewRequest("POST", "/cache", &mockRequestReader)

	// Run test
	router.ServeHTTP(recorder, request)

	// Assert
	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected a bad request status code from a malformed request")
}

func TestTooManyPutElements(t *testing.T) {
	// Test case: request with more than elements than put handler's max number of values
	putElements := []string{
		"{\"type\":\"json\",\"value\":true}",
		"{\"type\":\"xml\",\"value\":\"plain text\"}",
		"{\"type\":\"xml\",\"value\":\"2\"}",
	}
	reqBody := fmt.Sprintf("{\"puts\":[%s, %s, %s]}", putElements[0], putElements[1], putElements[2])

	//Set up server with capacity to handle less than putElements.size()
	backend := backends.NewMemoryBackend()
	router := httprouter.New()
	router.POST("/cache", NewPutHandler(backend, len(putElements)-1, true))

	_, httpTestRecorder := doMockPut(t, router, reqBody)
	assert.Equalf(t, http.StatusBadRequest, httpTestRecorder.Code, "doMockPut should have failed when trying to store %d elements because capacity is %d ", len(putElements), len(putElements)-1)
}

func TestMultiPutRequest(t *testing.T) {
	// Test case: request with more than one element in the "puts" array
	type aTest struct {
		description         string
		elemToPut           string
		expectedStoredValue string
	}
	testCases := []aTest{
		{
			description:         "Post in JSON format that contains a bool",
			elemToPut:           "{\"type\":\"json\",\"value\":true}",
			expectedStoredValue: "true",
		},
		{
			description:         "Post in XML format containing plain text",
			elemToPut:           "{\"type\":\"xml\",\"value\":\"plain text\"}",
			expectedStoredValue: "plain text",
		},
		{
			description:         "Post in XML format containing escaped double quotes",
			elemToPut:           "{\"type\":\"xml\",\"value\":\"2\"}",
			expectedStoredValue: "2",
		},
	}
	reqBody := fmt.Sprintf("{\"puts\":[%s, %s, %s]}", testCases[0].elemToPut, testCases[1].elemToPut, testCases[2].elemToPut)

	request, err := http.NewRequest("POST", "/cache", strings.NewReader(reqBody))
	assert.NoError(t, err, "Failed to create a POST request: %v", err)

	//Set up server and run
	router := httprouter.New()
	backend := backends.NewMemoryBackend()

	router.POST("/cache", NewPutHandler(backend, 10, true))
	router.GET("/cache", NewGetHandler(backend, true))

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, request)

	// validate results
	var parsed PutResponse
	err = json.Unmarshal([]byte(rr.Body.String()), &parsed)
	assert.NoError(t, err, "Response from POST doesn't conform to the expected format: %s", rr.Body.String())

	for i, resp := range parsed.Responses {
		// Get value for this UUID. It is supposed to have been stored
		getResult := doMockGet(t, router, resp.UUID)

		// Assertions
		assert.Equalf(t, http.StatusOK, getResult.Code, "Description: %s \n Multi-element put failed to store:%s \n", testCases[i].description, testCases[i].elemToPut)
		assert.Equalf(t, testCases[i].expectedStoredValue, getResult.Body.String(), "GET response error. Expected %v. Actual %v", testCases[i].expectedStoredValue, getResult.Body.String())
	}
}

func TestBadPayloadSizePutError(t *testing.T) {
	// Stored value size_limit
	sizeLimit := 3

	// Request with a string longer than sizeLimit
	reqBody := "{\"puts\":[{\"type\":\"xml\",\"value\":\"text longer than size limit\"}]}"

	// Declare a sizeCappedBackend client
	backend := backendDecorators.EnforceSizeLimit(backends.NewMemoryBackend(), sizeLimit)

	// Run client
	router := httprouter.New()
	router.POST("/cache", NewPutHandler(backend, 10, true))

	_, httpTestRecorder := doMockPut(t, router, reqBody)

	// Assert
	assert.Equal(t, http.StatusBadRequest, httpTestRecorder.Code, "doMockPut should have failed when trying to store elements in sizeCappedBackend")
}

func TestInternalPutClientError(t *testing.T) {
	// Valid request
	reqBody := "{\"puts\":[{\"type\":\"xml\",\"value\":\"text longer than size limit\"}]}"

	// Use mock client that will return an error
	backend := NewErrorReturningBackend()

	// Run client
	router := httprouter.New()
	router.POST("/cache", NewPutHandler(backend, 10, true))

	_, httpTestRecorder := doMockPut(t, router, reqBody)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, httpTestRecorder.Code, "Put should have failed because we are using an MockReturnErrorBackend")
}

func TestEmptyPutRequests(t *testing.T) {
	// Test case: request with more than one element in the "puts" array
	type aTest struct {
		description      string
		reqBody          string
		expectedResponse string
		emptyResponses   bool
	}
	testCases := []aTest{
		{
			description:      "Blank value in put element",
			reqBody:          "{\"puts\":[{\"type\":\"xml\",\"value\":\"\"}]}",
			expectedResponse: "{\"responses\":[\"uuid\":\"\"]}",
			emptyResponses:   false,
		},
		// This test is meant to come right after the "Blank value in put element" test in order to assert the correction
		// of a bug in the pre-PR#64 version of `endpoints/put.go`
		{
			description:      "All empty body. ",
			reqBody:          "{}",
			expectedResponse: "{\"responses\":[]}",
			emptyResponses:   true,
		},
		{
			description:      "Empty puts arrray",
			reqBody:          "{\"puts\":[]}",
			expectedResponse: "{\"responses\":[]}",
			emptyResponses:   true,
		},
	}

	// Set up server
	router := httprouter.New()
	backend := backends.NewMemoryBackend()

	router.POST("/cache", NewPutHandler(backend, 10, true))

	for i, test := range testCases {
		rr := httptest.NewRecorder()

		// Create request everytime
		request, err := http.NewRequest("POST", "/cache", strings.NewReader(test.reqBody))
		assert.NoError(t, err, "[%d] Failed to create a POST request: %v", i, err)

		// Run
		router.ServeHTTP(rr, request)
		assert.Equal(t, http.StatusOK, rr.Code, "[%d] ServeHTTP(rr, request) failed = %v \n", i, rr.Result())

		// validate results
		if test.emptyResponses && !assert.Equal(t, test.expectedResponse, rr.Body.String(), "[%d] Text response not empty as expected", i) {
			return
		}

		var parsed PutResponse
		err2 := json.Unmarshal([]byte(rr.Body.String()), &parsed)
		assert.NoError(t, err2, "[%d] Error found trying to unmarshal: %s \n", i, rr.Body.String())

		if test.emptyResponses {
			assert.Equal(t, 0, len(parsed.Responses), "[%d] This is NOT an empty response len(parsed.Responses) = %d; parsed.Responses = %v \n", i, len(parsed.Responses), parsed.Responses)
		} else {
			assert.Greater(t, len(parsed.Responses), 0, "[%d] This is an empty response len(parsed.Responses) = %d; parsed.Responses = %v \n", i, len(parsed.Responses), parsed.Responses)
		}
	}
}

func TestPutClientDeadlineExceeded(t *testing.T) {
	// Valid request
	reqBody := "{\"puts\":[{\"type\":\"xml\",\"value\":\"text longer than size limit\"}]}"

	// Use mock client that will return an error
	backend := NewDeadlineExceededBackend()

	// Run client
	router := httprouter.New()
	router.POST("/cache", NewPutHandler(backend, 10, true))

	_, httpTestRecorder := doMockPut(t, router, reqBody)

	// Assert
	assert.Equal(t, HttpDependencyTimeout, httpTestRecorder.Code, "Put should have failed because we are using a MockDeadlineExceededBackend")
}

func BenchmarkPutHandlerLen1(b *testing.B) {
	b.StopTimer()

	input := "{\"puts\":[{\"type\":\"json\",\"value\":\"plain text\"}]}"
	benchmarkPutHandler(b, input)
}

func BenchmarkPutHandlerLen2(b *testing.B) {
	b.StopTimer()

	//Set up a request that should succeed
	input := "{\"puts\":[{\"type\":\"json\",\"value\":true}, {\"type\":\"xml\",\"value\":\"plain text\"}]}"
	benchmarkPutHandler(b, input)
}

func BenchmarkPutHandlerLen4(b *testing.B) {
	b.StopTimer()

	//Set up a request that should succeed
	input := "{\"puts\":[{\"type\":\"json\",\"value\":true}, {\"type\":\"xml\",\"value\":\"plain text\"},{\"type\":\"xml\",\"value\":5}, {\"type\":\"json\",\"value\":\"esca\\\"ped\"}]}"
	benchmarkPutHandler(b, input)
}

func BenchmarkPutHandlerLen8(b *testing.B) {
	b.StopTimer()

	//Set up a request that should succeed
	input := "{\"puts\":[{\"type\":\"json\",\"value\":true}, {\"type\":\"xml\",\"value\":\"plain text\"},{\"type\":\"xml\",\"value\":5}, {\"type\":\"json\",\"value\":\"esca\\\"ped\"}, {\"type\":\"json\",\"value\":{\"custom_key\":\"foo\"}},{\"type\":\"xml\",\"value\":{\"custom_key\":\"foo\"}},{\"type\":\"json\",\"value\":null}, {\"type\":\"xml\",\"value\":\"<tag></tag>\"}]}"
	benchmarkPutHandler(b, input)
}

func doMockGet(t *testing.T, router *httprouter.Router, id string) *httptest.ResponseRecorder {
	requestRecorder := httptest.NewRecorder()

	body := new(bytes.Buffer)
	getReq, err := http.NewRequest("GET", "/cache"+"?uuid="+id, body)
	if err != nil {
		t.Fatalf("Failed to create a GET request: %v", err)
		return requestRecorder
	}
	router.ServeHTTP(requestRecorder, getReq)
	return requestRecorder
}

func doMockPut(t *testing.T, router *httprouter.Router, content string) (string, *httptest.ResponseRecorder) {
	var parseMockUUID = func(t *testing.T, putResponse string) string {
		var parsed PutResponse
		err := json.Unmarshal([]byte(putResponse), &parsed)
		if err != nil {
			t.Errorf("Response from POST doesn't conform to the expected format: %v", putResponse)
		}
		return parsed.Responses[0].UUID
	}

	rr := httptest.NewRecorder()

	request, err := http.NewRequest("POST", "/cache", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Failed to create a POST request: %v", err)
		return "", rr
	}

	router.ServeHTTP(rr, request)
	uuid := ""
	if rr.Code == http.StatusOK {
		uuid = parseMockUUID(t, rr.Body.String())
	}
	return uuid, rr
}

// expectStored makes a POST request with the given putBody, and then makes sure that expectedGet
// is returned by the GET request for whatever UUID the server chose.
func expectStored(t *testing.T, putBody string, expectedGet string, expectedMimeType string) {
	router := httprouter.New()
	backend := backends.NewMemoryBackend()

	router.POST("/cache", NewPutHandler(backend, 10, true))
	router.GET("/cache", NewGetHandler(backend, true))

	uuid, putTrace := doMockPut(t, router, putBody)
	if putTrace.Code != http.StatusOK {
		t.Fatalf("Put command failed. Status: %d, Msg: %v", putTrace.Code, putTrace.Body.String())
		return
	}

	getResults := doMockGet(t, router, uuid)
	if getResults.Code != http.StatusOK {
		t.Fatalf("Get command failed with status: %d", getResults.Code)
		return
	}
	if getResults.Body.String() != expectedGet {
		t.Fatalf("Expected GET response %v to equal %v", getResults.Body.String(), expectedGet)
		return
	}
	if getResults.Header().Get("Content-Type") != expectedMimeType {
		t.Fatalf("Expected GET response Content-Type %v to equal %v", getResults.Header().Get("Content-Type"), expectedMimeType)
	}
}

// expectFailedPut makes a POST request with the given request body, and fails unless the server
// responds with a 400
func expectFailedPut(t *testing.T, requestBody string) {
	backend := backends.NewMemoryBackend()
	router := httprouter.New()
	router.POST("/cache", NewPutHandler(backend, 10, true))

	_, putTrace := doMockPut(t, router, requestBody)
	if putTrace.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 response. Got: %d, Msg: %v", putTrace.Code, putTrace.Body.String())
		return
	}
}

func benchmarkPutHandler(b *testing.B, testCase string) {
	b.StopTimer()
	//Set up a request that should succeed
	request, err := http.NewRequest("POST", "/cache", strings.NewReader(testCase))
	if err != nil {
		b.Errorf("Failed to create a POST request: %v", err)
	}

	//Set up server ready to run
	router := httprouter.New()
	backend := backends.NewMemoryBackend()

	router.POST("/cache", NewPutHandler(backend, 10, true))
	router.GET("/cache", NewGetHandler(backend, true))

	rr := httptest.NewRecorder()

	//for statement to execute handler function
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		router.ServeHTTP(rr, request)
		b.StopTimer()
	}
}

func newMockBackend() *backends.MemoryBackend {
	backend := backends.NewMemoryBackend()

	backend.Put(context.TODO(), "non-36-char-key-maps-to-json", `json{"field":"value"}`, 0)
	backend.Put(context.TODO(), "36-char-key-maps-to-non-xml-nor-json", `#@!*{"desc":"data got malformed and is not prefixed with 'xml' nor 'json' substring"}`, 0)
	backend.Put(context.TODO(), "36-char-key-maps-to-actual-xml-value", "xml<tag>xml data here</tag>", 0)

	return backend
}

type faultyRequestBodyReader struct {
	mock.Mock
}

func (b *faultyRequestBodyReader) Read(p []byte) (n int, err error) {
	args := b.Called(p)
	return args.Int(0), args.Error(1)
}

func (b *faultyRequestBodyReader) Close() error {
	args := b.Called()
	return args.Error(0)
}

type errorReturningBackend struct{}

func (b *errorReturningBackend) Get(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("This is a mock backend that returns this error on Get() operation")
}

func (b *errorReturningBackend) Put(ctx context.Context, key string, value string, ttlSeconds int) error {
	return fmt.Errorf("This is a mock backend that returns this error on Put() operation")
}

func NewErrorReturningBackend() *errorReturningBackend {
	return &errorReturningBackend{}
}

type deadlineExceedingBackend struct{}

func (b *deadlineExceedingBackend) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (b *deadlineExceedingBackend) Put(ctx context.Context, key string, value string, ttlSeconds int) error {
	var err error

	d := time.Now().Add(50 * time.Millisecond)
	sampleCtx, cancel := context.WithDeadline(context.Background(), d)

	// Even though ctx will be expired, it is good practice to call its
	// cancellation function in any case. Failure to do so may keep the
	// context and its parent alive longer than necessary.
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		//err = fmt.Errorf("Some other error")
		err = nil
	case <-sampleCtx.Done():
		err = sampleCtx.Err()
	}
	return err
}

func NewDeadlineExceededBackend() *deadlineExceedingBackend {
	return &deadlineExceedingBackend{}
}

func TestHealthCheck(t *testing.T) {
	requestRecorder := httptest.NewRecorder()

	router := httprouter.New()
	router.GET("/healthcheck", HealthCheck)
	req, _ := http.NewRequest("GET", "/healthcheck", new(bytes.Buffer))
	router.ServeHTTP(requestRecorder, req)

	if requestRecorder.Code != http.StatusOK {
		t.Errorf("/healthcheck endpoint should always return a 200. Got %d", requestRecorder.Code)
	}
}
