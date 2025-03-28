package backends

import (
	"context"
	"fmt"
	"testing"

	as "github.com/aerospike/aerospike-client-go/v7"
	as_types "github.com/aerospike/aerospike-client-go/v7/types"
	"github.com/prebid/prebid-cache/metrics"
	"github.com/prebid/prebid-cache/metrics/metricstest"
	"github.com/prebid/prebid-cache/utils"
	"github.com/stretchr/testify/assert"
)

func TestClassifyAerospikeError(t *testing.T) {
	testCases := []struct {
		desc        string
		inErr       error
		expectedErr error
	}{
		{
			desc:        "Nil error",
			inErr:       nil,
			expectedErr: nil,
		},
		{
			desc:        "Generic non-nil error, expect same error in output",
			inErr:       fmt.Errorf("client.Get returned nil record"),
			expectedErr: fmt.Errorf("client.Get returned nil record"),
		},
		{
			desc:        "Aerospike error is neither KEY_NOT_FOUND_ERROR nor KEY_EXISTS_ERROR, expect same error as output",
			inErr:       &as.AerospikeError{ResultCode: as_types.SERVER_NOT_AVAILABLE},
			expectedErr: &as.AerospikeError{ResultCode: as_types.SERVER_NOT_AVAILABLE},
		},
		{
			desc:        "Aerospike KEY_NOT_FOUND_ERROR error, expect Prebid Cache's KEY_NOT_FOUND error",
			inErr:       &as.AerospikeError{ResultCode: as_types.KEY_NOT_FOUND_ERROR},
			expectedErr: utils.NewPBCError(utils.KEY_NOT_FOUND),
		},
		{
			desc:        "Aerospike KEY_EXISTS_ERROR error, expect Prebid Cache's RECORD_EXISTS error",
			inErr:       &as.AerospikeError{ResultCode: as_types.KEY_EXISTS_ERROR},
			expectedErr: utils.NewPBCError(utils.RECORD_EXISTS),
		},
	}
	for _, test := range testCases {
		actualErr := classifyAerospikeError(test.inErr)
		if test.expectedErr == nil {
			assert.Nil(t, actualErr, test.desc)
		} else {
			assert.Equal(t, test.expectedErr.Error(), actualErr.Error(), test.desc)
		}
	}
}

func TestAerospikeClientGet(t *testing.T) {
	mockMetrics := metricstest.CreateMockMetrics()
	m := &metrics.Metrics{
		MetricEngines: []metrics.CacheMetrics{
			&mockMetrics,
		},
	}
	aerospikeBackend := &AerospikeBackend{
		metrics: m,
	}

	testCases := []struct {
		desc              string
		inAerospikeClient AerospikeDB
		expectedValue     string
		expectedErrorMsg  string
	}{
		{
			desc:              "AerospikeBackend.Get() throws error when trying to generate new key",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_KEY_GEN_ERROR"},
			expectedValue:     "",
			expectedErrorMsg:  "ResultCode: NOT_AUTHENTICATED, Iteration: 0, InDoubt: false, Node: <nil>: ",
		},
		{
			desc:              "AerospikeBackend.Get() throws error when 'client.Get(..)' gets called",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_GET_ERROR"},
			expectedValue:     "",
			expectedErrorMsg:  "Key not found",
		},
		{
			desc:              "AerospikeBackend.Get() throws error when 'client.Get(..)' returns a nil record",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_NIL_RECORD_ERROR"},
			expectedValue:     "",
			expectedErrorMsg:  "Nil record",
		},
		{
			desc:              "AerospikeBackend.Get() throws error no BIN_VALUE bucket is found",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_NO_BUCKET_ERROR"},
			expectedValue:     "",
			expectedErrorMsg:  "No 'value' bucket found",
		},
		{
			desc:              "AerospikeBackend.Get() returns a record that does not store a string",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_NON_STRING_VALUE_ERROR"},
			expectedValue:     "",
			expectedErrorMsg:  "Unexpected non-string value found",
		},
		{
			desc: "AerospikeBackend.Get() does not throw error",
			inAerospikeClient: &GoodAerospikeClient{
				StoredData: map[string]string{"defaultKey": "Default value"},
			},
			expectedValue:    "Default value",
			expectedErrorMsg: "",
		},
	}

	for _, tt := range testCases {
		// Assign aerospike backend cient
		aerospikeBackend.client = tt.inAerospikeClient

		// Run test
		actualValue, actualErr := aerospikeBackend.Get(context.Background(), "defaultKey")

		// Assertions
		assert.Equal(t, tt.expectedValue, actualValue, tt.desc)

		if tt.expectedErrorMsg == "" {
			assert.Nil(t, actualErr, tt.desc)
		} else {
			assert.Equal(t, tt.expectedErrorMsg, actualErr.Error(), tt.desc)
		}
	}
}

func TestClientPut(t *testing.T) {
	mockMetrics := metricstest.CreateMockMetrics()
	m := &metrics.Metrics{
		MetricEngines: []metrics.CacheMetrics{
			&mockMetrics,
		},
	}
	aerospikeBackend := &AerospikeBackend{
		metrics: m,
	}

	testCases := []struct {
		desc              string
		inAerospikeClient AerospikeDB
		inKey             string
		inValueToStore    string
		expectedStoredVal string
		expectedErrorMsg  string
	}{
		{
			desc:              "AerospikeBackend.Put() throws error when trying to generate new key",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_KEY_GEN_ERROR"},
			inKey:             "testKey",
			inValueToStore:    "not default value",
			expectedStoredVal: "",
			expectedErrorMsg:  "ResultCode: NOT_AUTHENTICATED, Iteration: 0, InDoubt: false, Node: <nil>: ",
		},
		{
			desc:              "AerospikeBackend.Put() throws error when 'client.Put(..)' gets called",
			inAerospikeClient: &ErrorProneAerospikeClient{ServerError: "TEST_PUT_ERROR"},
			inKey:             "testKey",
			inValueToStore:    "not default value",
			expectedStoredVal: "",
			expectedErrorMsg:  "Record exists with provided key.",
		},
		{
			desc: "AerospikeBackend.Put() does not throw error",
			inAerospikeClient: &GoodAerospikeClient{
				StoredData: map[string]string{"defaultKey": "Default value"},
			},
			inKey:             "testKey",
			inValueToStore:    "any value",
			expectedStoredVal: "any value",
			expectedErrorMsg:  "",
		},
	}

	for _, tt := range testCases {
		// Assign aerospike backend cient
		aerospikeBackend.client = tt.inAerospikeClient

		// Run test
		actualErr := aerospikeBackend.Put(context.Background(), tt.inKey, tt.inValueToStore, 0)

		// Assert Put error
		if tt.expectedErrorMsg != "" {
			assert.Equal(t, tt.expectedErrorMsg, actualErr.Error(), tt.desc)
		} else {
			assert.Nil(t, actualErr, tt.desc)

			// Assert Put() sucessfully logged "not default value" under "testKey":
			storedValue, getErr := aerospikeBackend.Get(context.Background(), tt.inKey)

			assert.Nil(t, getErr, tt.desc)
			assert.Equal(t, tt.inValueToStore, storedValue, tt.desc)
		}
	}
}
