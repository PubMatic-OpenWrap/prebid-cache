package endpoints

import (
	"net/http"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/julienschmidt/httprouter"
)

/*HealthCheck end-point for prebid-cache*/
func HealthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger.Debug("Health Check Request")
	w.WriteHeader(http.StatusOK)
	return
}
