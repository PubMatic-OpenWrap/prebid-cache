package compression

import (
	"context"
	"time"

	"git.pubmatic.com/PubMatic/go-common.git/logger"

	"github.com/golang/snappy"
	"github.com/prebid/prebid-cache/backends"
)

// SnappyCompress runs snappy compression on data before saving it in the backend.
// For more info, see https://en.wikipedia.org/wiki/Snappy_(compression)
func SnappyCompress(backend backends.Backend) backends.Backend {
	return &snappyCompressor{
		delegate: backend,
	}
}

type snappyCompressor struct {
	delegate backends.Backend
}

func (s *snappyCompressor) Put(ctx context.Context, key string, value string, rqID string) error {
	start := time.Now()

	p := s.delegate.Put(ctx, key, string(snappy.Encode(nil, []byte(value))), rqID)
	end := time.Now()
	totalTime := (end.Sub(start)).Nanoseconds() / 1000000
	logger.Info("Time for snappy put: %v, rqID: %s", totalTime, rqID)
	return p
}

func (s *snappyCompressor) Get(ctx context.Context, key string, rqID string) (string, error) {
	start := time.Now()
	compressed, err := s.delegate.Get(ctx, key, rqID)
	if err != nil {
		return "", err
	}

	decompressed, err := snappy.Decode(nil, []byte(compressed))
	if err != nil {
		return "", err
	}
	end := time.Now()
	totalTime := (end.Sub(start)).Nanoseconds() / 1000000
	logger.Info("Time for snappy get: %v, rqID: %s", totalTime, rqID)

	return string(decompressed), nil
}
