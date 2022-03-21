package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.pubmatic.com/PubMatic/go-common.git/logger"
	"github.com/PubMatic-OpenWrap/prebid-cache/backends"
	"github.com/PubMatic-OpenWrap/prebid-cache/constant"
	"github.com/PubMatic-OpenWrap/prebid-cache/metrics"
	"github.com/PubMatic-OpenWrap/prebid-cache/stats"
	"github.com/PubMatic-OpenWrap/prebid-cache/utils"
	"github.com/julienschmidt/httprouter"
)

// GetHandler serves "GET /cache" requests.
type GetHandler struct {
	backend         backends.Backend
	metrics         *metrics.Metrics
	allowCustomKeys bool
}

// NewGetHandler returns the handle function for the "/cache" endpoint when it gets receives a GET request
func NewGetHandler(storage backends.Backend, metrics *metrics.Metrics, allowCustomKeys bool) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	getHandler := &GetHandler{
		// Assign storage client to get endpoint
		backend: storage,
		// pass metrics engine
		metrics: metrics,
		// Pass configuration value
		allowCustomKeys: allowCustomKeys,
	}

	// Return handle function
	return getHandler.handle
}

func (e *GetHandler) handle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	e.metrics.RecordGetTotal()
	start := time.Now()
	logger.Info("Get /cache called")
	stats.LogCacheRequestedGetStats()

	uuid, parseErr := parseUUID(r, e.allowCustomKeys)
	if parseErr != nil {
		// parseUUID either returns http.StatusBadRequest or http.StatusNotFound. Both should be
		// accounted using RecordGetBadRequest()
		e.handleException(w, uuid, parseErr)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	storedData, err := e.backend.Get(ctx, uuid)
	if err != nil {
		stats.LogCacheMissStats()
		logger.Info("Cache miss for uuid: %v", uuid)
		e.handleException(w, uuid, err)
		logger.Info("Total time for get: %v", time.Now().Sub(start))
		return
	}

	if err := writeGetResponse(w, storedData); err != nil {
		e.handleException(w, uuid, err)
		return
	}

	// successfully retrieved value under uuid from the backend storage
	logger.Info("Total time for get: %v", time.Now().Sub(start))
	e.metrics.RecordGetDuration(time.Since(start))
	return
}

// parseUUID extracts the uuid value from the query and validates its
// lenght in case custom keys are not allowed.
func parseUUID(r *http.Request, allowCustomKeys bool) (string, error) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		stats.LogCacheFailedGetStats(constant.UUIDMissing)
		return "", utils.NewPBCError(utils.MISSING_KEY)
	}
	// UUIDs are 36 characters long... so this quick check lets us filter out most invalid
	// ones before even checking the backend.
	if len(uuid) != 36 && (!allowCustomKeys) {
		stats.LogCacheFailedGetStats(constant.InvalidUUID)
		return uuid, utils.NewPBCError(utils.KEY_LENGTH)
	}
	return uuid, nil
}

// writeGetResponse writes the "Content-Type" header and sends back the stored data as a response if
// the sotred data is prefixed by either the "xml" or "json"
func writeGetResponse(w http.ResponseWriter, storedData string) error {
	if strings.HasPrefix(storedData, backends.XML_PREFIX) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(storedData)[len(backends.XML_PREFIX):])
	} else if strings.HasPrefix(storedData, backends.JSON_PREFIX) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(storedData)[len(backends.JSON_PREFIX):])
	} else {
		return utils.NewPBCError(utils.UNKNOWN_STORED_DATA_TYPE)
	}
	return nil
}

// handleException logs the error message, updates the error metrics based on error type and replies
// back with the error message and an HTTP error code
func (e *GetHandler) handleException(w http.ResponseWriter, uuid string, err error) {
	if err != nil {
		// Prefix error message with "GET /cache " or "GET /cache uuid=..."
		errMsgBuilder := strings.Builder{}
		errMsgBuilder.WriteString("GET /cache")
		if len(uuid) > 0 {
			errMsgBuilder.WriteString(fmt.Sprintf(" uuid=%s", uuid))
		}
		errMsgBuilder.WriteString(fmt.Sprintf(": %s", err.Error()))
		errMsg := errMsgBuilder.String()

		// Determine the response status code based on error type
		errCode := http.StatusInternalServerError
		isKeyNotFound := false
		if pbcErr, isPBCErr := err.(utils.PBCError); isPBCErr {
			errCode = pbcErr.StatusCode
			isKeyNotFound = pbcErr.Type == utils.KEY_NOT_FOUND
		}

		// Log error metrics based on error type
		switch {
		case errCode >= http.StatusInternalServerError: // 500
			e.metrics.RecordGetError()
		case errCode >= http.StatusBadRequest: // 400
			e.metrics.RecordGetBadRequest()
		}

		// Determine log level
		if isKeyNotFound {
			logger.Debug(errMsg)
		} else {
			logger.Error(errMsg)
		}

		// Write error response
		http.Error(w, errMsg, errCode)
	}
}
