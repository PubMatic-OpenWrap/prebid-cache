package endpoints

import (
	"encoding/json"
	"net/http"

	"git.pubmatic.com/PubMatic/go-common.git/logger"
	"github.com/julienschmidt/httprouter"
)

const versionEndpointValueNotSet = "not-set"

// NewVersionEndpoint returns the latest git tag as the version and commit hash as the revision from which the binary was built
func NewVersionEndpoint(version, revision string) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	if version == "" {
		version = versionEndpointValueNotSet
	}
	if revision == "" {
		revision = versionEndpointValueNotSet
	}

	response, err := json.Marshal(struct {
		Revision string `json:"revision"`
		Version  string `json:"version"`
	}{
		Revision: revision,
		Version:  version,
	})
	if err != nil {
		logger.Error("error creating /version endpoint response: %v", err)
	}

	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Write(response)
	}
}
