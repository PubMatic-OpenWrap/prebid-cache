package backends

import (
	"context"
	"errors"
	"time"

	"git.pubmatic.com/PubMatic/go-common.git/logger"
	as "github.com/aerospike/aerospike-client-go"
	as_types "github.com/aerospike/aerospike-client-go/types"
	"github.com/prebid/prebid-cache/config"
	"github.com/prebid/prebid-cache/metrics"
	"github.com/prebid/prebid-cache/stats"
	"github.com/prebid/prebid-cache/utils"
)

const setName = "uuid"
const binValue = "value"

// AerospikeDB is a wrapper for the Aerospike client
type AerospikeDB interface {
	NewUUIDKey(namespace string, key string) (*as.Key, error)
	Get(key *as.Key) (*as.Record, error)
	Put(policy *as.WritePolicy, key *as.Key, binMap as.BinMap) error
}

// AerospikeDBClient implements the AerospikeDB interface
type AerospikeDBClient struct {
	client *as.Client
}

// Get performs the as.Client Get operation
func (db AerospikeDBClient) Get(key *as.Key) (*as.Record, error) {
	return db.client.Get(nil, key, binValue)
}

// Put performs the as.Client Put operation
func (db AerospikeDBClient) Put(policy *as.WritePolicy, key *as.Key, binMap as.BinMap) error {
	return db.client.Put(policy, key, binMap)
}

// NewUUIDKey creates an aerospike key so we can store data under it
func (db *AerospikeDBClient) NewUUIDKey(namespace string, key string) (*as.Key, error) {
	return as.NewKey(namespace, setName, key)
}

// AerospikeBackend upon creation will instantiates, and configure the Aerospike client. Implements
// the Backend interface
type AerospikeBackend struct {
	namespace string
	client    AerospikeDB
	metrics   *metrics.Metrics
}

// NewAerospikeBackend validates config.Aerospike and returns an AerospikeBackend
func NewAerospikeBackend(cfg config.Aerospike, metrics *metrics.Metrics) *AerospikeBackend {
	var hosts []*as.Host

	clientPolicy := as.NewClientPolicy()
	// cfg.User and cfg.Password are optional parameters
	// if left blank in the config, they will default to the empty
	// string and be ignored
	clientPolicy.User = cfg.User
	clientPolicy.Password = cfg.Password

	if len(cfg.Host) > 1 {
		hosts = append(hosts, as.NewHost(cfg.Host, cfg.Port))
		logger.Info("config.backend.aerospike.host is being deprecated in favor of config.backend.aerospike.hosts")
	}
	for _, host := range cfg.Hosts {
		hosts = append(hosts, as.NewHost(host, cfg.Port))
	}

	client, err := as.NewClientWithPolicyAndHost(clientPolicy, hosts...)
	if err != nil {
		stats.LogAerospikeErrorStats()
		logger.Fatal("Error creating Aerospike backend: %+v", err)
		panic("AerospikeBackend failure. This shouldn't happen.")
	}
	logger.Info("Connected to Aerospike host(s) %v on port %d", append(cfg.Hosts, cfg.Host), cfg.Port)

	return &AerospikeBackend{
		namespace: cfg.Namespace,
		client:    &AerospikeDBClient{client},
		metrics:   metrics,
	}
}

// Get creates an aerospike key based on the UUID key parameter, perfomrs the client's Get call
// and validates results. Can return a KEY_NOT_FOUND error or other Aerospike server errors
func (a *AerospikeBackend) Get(ctx context.Context, key string) (string, error) {
	aerospikeStartTime := time.Now()
	asKey, err := a.client.NewUUIDKey(a.namespace, key)
	if err != nil {
		return "", classifyAerospikeError(err)
	}
	rec, err := a.client.Get(asKey)
	if err != nil {
		return "", classifyAerospikeError(err)
	}
	if rec == nil {
		return "", errors.New("Nil record")
	}

	value, found := rec.Bins[binValue]
	if !found {
		return "", errors.New("No 'value' bucket found")
	}
	logger.Info("Time taken by Aerospike for get: %v", time.Now().Sub(aerospikeStartTime))

	str, isString := value.(string)
	if !isString {
		return "", errors.New("Unexpected non-string value found")
	}

	return str, nil
}

// Put creates an aerospike key based on the UUID key parameter and stores the value using the
// client's Put implementaion. Can return a RECORD_EXISTS error or other Aerospike server errors
func (a *AerospikeBackend) Put(ctx context.Context, key string, value string, ttlSeconds int) error {
	aerospikeStartTime := time.Now()
	asKey, err := a.client.NewUUIDKey(a.namespace, key)
	if err != nil {
		return classifyAerospikeError(err)
	}

	bins := as.BinMap{binValue: value}
	policy := &as.WritePolicy{
		Expiration:         uint32(ttlSeconds),
		RecordExistsAction: as.CREATE_ONLY,
	}

	if err := a.client.Put(policy, asKey, bins); err != nil {
		return classifyAerospikeError(err)
	}

	logger.Info("Time taken by Aerospike for put: %v", time.Now().Sub(aerospikeStartTime))
	return nil
}

func classifyAerospikeError(err error) error {
	if err != nil {
		if aerr, ok := err.(as_types.AerospikeError); ok {
			if aerr.ResultCode() == as_types.KEY_NOT_FOUND_ERROR {
				return utils.NewPBCError(utils.KEY_NOT_FOUND)
			}
			if aerr.ResultCode() == as_types.KEY_EXISTS_ERROR {
				return utils.NewPBCError(utils.RECORD_EXISTS)
			}
		}
	}
	return err
}
