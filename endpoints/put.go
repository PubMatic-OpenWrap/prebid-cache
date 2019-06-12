package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.pubmatic.com/PubMatic/go-common.git/logger"
	"github.com/PubMatic-OpenWrap/prebid-cache/backends"
	"github.com/PubMatic-OpenWrap/prebid-cache/constant"
	"github.com/PubMatic-OpenWrap/prebid-cache/stats"
	"github.com/julienschmidt/httprouter"
)

// PutHandler serves "POST /cache" requests.
func NewPutHandler(backend backends.Backend, maxNumValues int) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	// TODO(future PR): Break this giant function apart
	putAnyRequestPool := sync.Pool{
		New: func() interface{} {
			//return PutRequest{}
			return []ReqObj{}
		},
	}

	putResponsePool := sync.Pool{
		New: func() interface{} {
			//return PutResponse{}
			return []string{}
		},
	}

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		start := time.Now()
		stats.LogCacheRequestedPutStats()
		logger.Info("POST /cache called")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error("Failed to read the request body.")
			http.Error(w, "Failed to read the request body.", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		put := putAnyRequestPool.Get().([]ReqObj)
		defer putAnyRequestPool.Put(put)

		err = json.Unmarshal(body, &put)
		if err != nil {
			stats.LogCacheFailedPutStats(constant.InvalidJSON)
			http.Error(w, "Request body "+string(body)+" is not valid JSON.", http.StatusBadRequest)
			return
		}

		/*if len(put.Puts) > maxNumValues {
			stats.LogCacheFailedPutStats(constant.KeyCountExceeded)
			http.Error(w, fmt.Sprintf("More keys than allowed: %d", maxNumValues), http.StatusBadRequest)
			return
		}*/

		resps := putResponsePool.Get().([]string)
		//resps.BlockedCreativeIds = make([]string, len(put))
		defer putResponsePool.Put(resps)

		for _, p := range put {
			/*if len(p.Value) == 0 {
				logger.Error("Missing value")
				http.Error(w, "Missing value.", http.StatusBadRequest)
				return
			}*/

			var toCache string
			/*if p.Type == backends.XML_PREFIX {
				if p.Value[0] != byte('"') || p.Value[len(p.Value)-1] != byte('"') {
					logger.Error("XML messages must have a String value. Found %v", p.Value)
					http.Error(w, fmt.Sprintf("XML messages must have a String value. Found %v", p.Value), http.StatusBadRequest)
					return
				}

				// Be careful about the the cross-script escaping issues here. JSON requires quotation marks to be escaped,
				// for example... so we'll need to un-escape it before we consider it to be XML content.
				var interpreted string
				json.Unmarshal(p.Value, &interpreted)
				toCache = p.Type + interpreted
			} else if p.Type == backends.JSON_PREFIX {
				toCache = p.Type + string(p.Value)
			} else {
				logger.Error("Type must be one of [\"json\", \"xml\"]. Found %v", p.Type)
				http.Error(w, fmt.Sprintf("Type must be one of [\"json\", \"xml\"]. Found %v", p.Type), http.StatusBadRequest)
				return
			}*/

			logger.Debug("Storing value: %s", toCache)

			ucrid := p.CreativeId + p.PartnerName

			/*var aqObj AqObject
			err = json.Unmarshal([]byte(p.Value), &aqObj)
			fmt.Println("object to save:", p.Value)
			fmt.Println("object to save:", aqObj)
			fmt.Println("error is :", err)
			resps.Responses[i].Ucrid = aqObj.Ucrid*/

			//resps.Responses[i].UUID = uuid.NewV4().String()
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			backendStartTime := time.Now()
			//err = backend.Put(ctx, resps.Responses[i].Ucrid, aqObj.IsMalware)
			value, err := backend.Get(ctx, ucrid)
			fmt.Println("ucrid:", ucrid)
			//fmt.Println("value:", (value)[len(backends.JSON_PREFIX):])
			fmt.Println("error is :", err)
			backendEndTime := time.Now()
			backendDiffTime := (backendEndTime.Sub(backendStartTime)).Nanoseconds() / 1000000
			logger.Info("Time taken by backend.Put: %v", backendDiffTime)
			/*if err != nil {

				if _, ok := err.(*backendDecorators.BadPayloadSize); ok {
					stats.LogCacheFailedPutStats(constant.MaxSizeExceeded)
					http.Error(w, fmt.Sprintf("POST /cache element %d exceeded max size: %v", i, err), http.StatusBadRequest)
					return
				}

				logger.Error("POST /cache Error while writing to the backend: ", err)
				switch err {
				case context.DeadlineExceeded:
					stats.LogCacheFailedPutStats(constant.TimedOut)
					logger.Error("POST /cache timed out:", err)
					http.Error(w, "Timeout writing value to the backend", HttpDependencyTimeout)
				default:
					stats.LogCacheFailedPutStats(constant.UnexpErr)
					logger.Error("POST /cache had an unexpected error:", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				end := time.Now()
				totalTime := (end.Sub(start)).Nanoseconds() / 1000000
				logger.Info("Total time for put: %v", totalTime)
				return
			}*/
			// log info
			//bid := make(map[string]interface{})
			fmt.Println("Error is:", err)
			if err == nil {

				//var bodyStr string
				if strings.HasPrefix(value, backends.JSON_PREFIX) {
					value = value[len(backends.JSON_PREFIX):]
				}
				//fmt.Println("jsonPrefixLen : ", jsonPrefixLen)
				fmt.Println("creativeId len:", len(p.CreativeId))
				fmt.Println("Value", value)

				//json.Unmarshal([]byte(value), &bodyStr)
				if value == "true" {
					fmt.Println("Inside value = true")
					resps = append(resps, ucrid[:len(p.CreativeId)])
				}
			}
			/*bodyByte := []byte(bodyStr)
			json.Unmarshal(bodyByte, &bid)
			if bid != nil && bid["ext"] != nil {
				bidExt := bid["ext"].(map[string]interface{})
				pubID := bidExt["pubId"]           // TODO: check key name and type
				platformID := bidExt["platformId"] // TODO: check key name and type
				requestID := bidExt["requestId"]   // TODO: check key name and type
				log.DebugWithRequestID(requestID.(string), "pubId: %s, platformId: %s, UUID: %s, Time: %v, Referer: %s", pubID, platformID, resps.Responses[i].Ucrid, start.Unix(), r.Referer())
			}*/

		}
		/*if resps.BlockedCreativeIds == nil || len(resps.BlockedCreativeIds) == 0 {
			resps.BlockedCreativeIds[0] = ""
		}*/

		bytes, err := json.Marshal(&resps)
		if err != nil {
			logger.Error("Failed to serialize UUIDs into JSON.")
			http.Error(w, "Failed to serialize UUIDs into JSON.", http.StatusInternalServerError)
			return
		}

		/* Handles POST */
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
		end := time.Now()
		totalTime := (end.Sub(start)).Nanoseconds() / 1000000
		logger.Info("Total time for put: %v", totalTime)
	}
}

type PutRequest struct {
	//Puts []PutObject `json:"puts"`
	Puts []ReqObj `json:"puts"`
}

type ReqObj struct {
	CreativeId  string `json:"creativeId"`
	Creative    string `json:"creative"`
	PartnerName string `json:"partnerName"`
}

type PutObject struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type PutResponseObject struct {
	Ucrid string `json:"ucrid"`
}

type PutResponse struct {
	BlockedCreativeIds []string `json:"blockedCreativeIds"`
}

type AqObject struct {
	Ucrid     string `json:"ucrid"`
	IsMalware string `json:"isMalware"`
}
