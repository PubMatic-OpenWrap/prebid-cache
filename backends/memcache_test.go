package backends

import (
	"context"
	"errors"
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-cache/utils"
	"github.com/bradfitz/gomemcache/memcache"
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
				&errorProneMemcache{errorToThrow: memcache.ErrCacheMiss},
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
				&errorProneMemcache{errorToThrow: errors.New("some other get error")},
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
				&goodMemcache{key: "defaultKey", value: "aValue"},
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
		actualValue, actualErr := mcBackend.Get(context.TODO(), tt.in.key)

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
				&errorProneMemcache{errorToThrow: memcache.ErrServerError},
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
				&errorProneMemcache{errorToThrow: memcache.ErrNotStored},
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
				&goodMemcache{key: "defaultKey", value: "aValue"},
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
		actualErr := mcBackend.Put(context.TODO(), tt.in.key, tt.in.valueToStore, tt.in.ttl)

		// Assert Put error
		assert.Equal(t, tt.expected.err, actualErr, tt.desc)

		// Assert value
		if tt.expected.err == nil {
			storedValue, getErr := mcBackend.Get(context.TODO(), tt.in.key)

			assert.NoError(t, getErr, tt.desc)
			assert.Equal(t, tt.expected.value, storedValue, tt.desc)
		}
	}
}

// Memcache that always throws an error
type errorProneMemcache struct {
	errorToThrow error
}

func (ec *errorProneMemcache) Get(key string) (*memcache.Item, error) {
	return nil, ec.errorToThrow
}

func (ec *errorProneMemcache) Put(key string, value string, ttlSeconds int) error {
	return ec.errorToThrow
}

// Memcache client that does not throw errors
type goodMemcache struct {
	key   string
	value string
}

func (gc *goodMemcache) Get(key string) (*memcache.Item, error) {
	if key == gc.key {
		return &memcache.Item{Key: gc.key, Value: []byte(gc.value)}, nil
	}
	return nil, utils.NewPBCError(utils.KEY_NOT_FOUND)
}

func (gc *goodMemcache) Put(key string, value string, ttlSeconds int) error {
	if gc.key != key {
		gc.key = key
	}
	gc.value = value

	return nil
}
