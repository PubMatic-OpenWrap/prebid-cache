package backends

import (
	"context"
	"errors"
	"testing"

	"github.com/google/gomemcache/memcache"
	"github.com/prebid/prebid-cache/utils"
	"github.com/stretchr/testify/assert"
)

func TestMemcacheGet(t *testing.T) {
	mcBackend := &MemcacheBackend{}

	type testInput struct {
		memcacheClient MemcacheDataStore
		key            string
	}

	type testExpectedValues struct {
		value string
		err   error
	}

	testCases := []struct {
		desc     string
		in       testInput
		expected testExpectedValues
	}{
		{
			"Memcache.Get() throws a memcache.ErrCacheMiss error",
			testInput{
				&ErrorProneMemcache{ServerError: memcache.ErrCacheMiss},
				"someKeyThatWontBeFound",
			},
			testExpectedValues{
				value: "",
				err:   utils.NewPBCError(utils.KEY_NOT_FOUND),
			},
		},
		{
			"Memcache.Get() throws an error different from Cassandra ErrNotFound error",
			testInput{
				&ErrorProneMemcache{ServerError: errors.New("some other get error")},
				"someKey",
			},
			testExpectedValues{
				value: "",
				err:   errors.New("some other get error"),
			},
		},
		{
			"Memcache.Get() doesn't throw an error",
			testInput{
				&GoodMemcache{StoredData: map[string]string{"defaultKey": "aValue"}},
				"defaultKey",
			},
			testExpectedValues{
				value: "aValue",
				err:   nil,
			},
		},
	}

	for _, tt := range testCases {
		mcBackend.memcache = tt.in.memcacheClient

		// Run test
		actualValue, actualErr := mcBackend.Get(context.Background(), tt.in.key)

		// Assertions
		assert.Equal(t, tt.expected.value, actualValue, tt.desc)
		assert.Equal(t, tt.expected.err, actualErr, tt.desc)
	}
}

func TestMemcachePut(t *testing.T) {
	mcBackend := &MemcacheBackend{}

	type testInput struct {
		memcacheClient MemcacheDataStore
		key            string
		valueToStore   string
		ttl            int
	}

	type testExpectedValues struct {
		value string
		err   error
	}

	testCases := []struct {
		desc     string
		in       testInput
		expected testExpectedValues
	}{
		{
			"Memcache.Put() throws non-ErrNotStored error",
			testInput{
				&ErrorProneMemcache{ServerError: memcache.ErrServerError},
				"someKey",
				"someValue",
				10,
			},
			testExpectedValues{
				"",
				memcache.ErrServerError,
			},
		},
		{
			"Memcache.Put() throws ErrNotStored error",
			testInput{
				&ErrorProneMemcache{ServerError: memcache.ErrNotStored},
				"someKey",
				"someValue",
				10,
			},
			testExpectedValues{
				"",
				utils.NewPBCError(utils.RECORD_EXISTS),
			},
		},
		{
			"Memcache.Put() successful",
			testInput{
				&GoodMemcache{StoredData: map[string]string{"defaultKey": "aValue"}},
				"defaultKey",
				"aValue",
				1,
			},
			testExpectedValues{
				"aValue",
				nil,
			},
		},
	}

	for _, tt := range testCases {
		mcBackend.memcache = tt.in.memcacheClient

		// Run test
		actualErr := mcBackend.Put(context.Background(), tt.in.key, tt.in.valueToStore, tt.in.ttl)

		// Assert Put error
		assert.Equal(t, tt.expected.err, actualErr, tt.desc)

		// Assert value
		if tt.expected.err == nil {
			storedValue, getErr := mcBackend.Get(context.Background(), tt.in.key)

			assert.NoError(t, getErr, tt.desc)
			assert.Equal(t, tt.expected.value, storedValue, tt.desc)
		}
	}
}
