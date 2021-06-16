package kvserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"git.pubmatic.com/PubMatic/go-common.git/logger"
	"github.com/julienschmidt/httprouter"
)

const SUCCESS = `{"status":"success"}`
const ERROR_RESPONSE = `{"endpoint":"%s", "error":"%v"}`

//InitKVServerHandlers initialise all handlers
func InitKVServerHandlers(router *httprouter.Router) {

	//lineitems
	router.GET("/kv/add/lineitem", kv_add_lineitem)
	router.GET("/kv/list/lineitem", kv_list_lineitem)
	//router.GET("/kv/edit/lineitem", kv_edit_lineitem)
	//router.GET("/kv/del/lineitem", kv_del_lineitem)

	//creatives
	router.GET("/kv/add/creative", kv_add_creative)
	router.GET("/kv/list/creative", kv_list_creative)
	//router.GET("/kv/edit/creative", kv_edit_creative)
	//router.GET("/kv/del/creative", kv_del_creative)

	//lineitem_creative_mapping
	router.GET("/kv/add/licr", kv_add_licr)
	router.GET("/kv/del/licr", kv_del_licr)
	//router.GET("/kv/list/licr", kv_list_licr)

	//contextual_signal_interest_group_mapping
	//router.GET("/kv/add/csig", kv_add_csig)
	//router.GET("/kv/del/csig", kv_del_csig)
	router.GET("/kv/get/csig", kv_get_csig)

	//impression_count
	router.POST("/kv/update/ic", kv_update_ic)
	router.GET("/kv/flush", flush)
}

func default_error_response(w http.ResponseWriter, r *http.Request, err error) {
	w.Write([]byte(fmt.Sprintf(ERROR_RESPONSE, r.URL.Path, err.Error())))
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(http.StatusBadRequest)
}

/*success_response end-point for prebid-cache*/
func success_response(w http.ResponseWriter, r *http.Request, response string) {
	w.Write([]byte(response))
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(http.StatusOK)
}

//lineitems
func kv_add_lineitem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger.Debug("Adding new lineitem")

	li, err := NewLineItem(
		GetInt(r.FormValue("id")),
		GetInt(r.FormValue("type")),
		GetFloat64(r.FormValue("price")),
		r.FormValue("source"),
		r.FormValue("device"),
		r.FormValue("os"),
		r.FormValue("ig"),
		r.FormValue("fcap"),
		r.FormValue("startdate"), r.FormValue("enddate"),
		GetInt(r.FormValue("goal")))
	if err != nil {
		default_error_response(w, r, err)
		return
	}

	//Add New LineItem
	AddNewLineItem(li)

	//Adding Mapping
	//AddCSIGLIMapping(li.RegExKey, li.ID)

	success_response(w, r, SUCCESS)
}
func kv_list_lineitem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var response []byte
	if len(r.FormValue("id")) == 0 {
		response, _ = json.Marshal(lineItemMap)
	} else {
		id := GetInt(r.FormValue("id"))
		li, ok := lineItemMap[id]
		if !ok {
			default_error_response(w, r, fmt.Errorf("lineitem not present : %d", id))
			return
		}
		response, _ = json.Marshal(li)
	}

	success_response(w, r, string(response))
}

//creatives
func kv_add_creative(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger.Debug("Adding new creative")

	cr, err := NewCreative(
		GetInt(r.FormValue("id")),
		r.FormValue("adm"),
		r.FormValue("type"))
	if err != nil {
		default_error_response(w, r, err)
		return
	}

	AddNewCreative(cr)
	success_response(w, r, SUCCESS)
}
func kv_list_creative(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var response []byte
	if len(r.FormValue("id")) == 0 {
		response, _ = json.Marshal(creativeMap)
	} else {
		id := GetInt(r.FormValue("id"))
		li, ok := creativeMap[id]
		if !ok {
			default_error_response(w, r, fmt.Errorf("creative not present : %d", id))
			return
		}
		response, _ = json.Marshal(li)
	}

	success_response(w, r, string(response))
}

//lineitem_creative_mapping
func kv_add_licr(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger.Debug("Mapping new lineitem creative mapping")
	liid := GetInt(r.FormValue("liid"))
	creatives := strings.Split(r.FormValue("crid"), ",")

	if _, ok := lineItemMap[liid]; !ok {
		default_error_response(w, r, fmt.Errorf("lineitem not present : %d", liid))
		return
	}

	for _, creative := range creatives {
		crid := GetInt(creative)
		if _, ok := creativeMap[crid]; ok {
			AddNewLineItemCreativeMapping(liid, crid)
		}
	}

	success_response(w, r, SUCCESS)
}
func kv_del_licr(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger.Debug("Unmapping new lineitem creative mapping")
	liid := GetInt(r.FormValue("liid"))
	creatives := strings.Split(r.FormValue("crid"), ",")

	if _, ok := lineItemMap[liid]; !ok {
		default_error_response(w, r, fmt.Errorf("lineitem not present : %d", liid))
		return
	}

	for _, creative := range creatives {
		crid := GetInt(creative)
		if _, ok := creativeMap[crid]; ok {
			UnmapLineItemCreativeMapping(liid, crid)
		}
	}

	success_response(w, r, SUCCESS)
}

//contextual_signal_interest_group_mapping
/*
func kv_add_csig(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cskey := r.FormValue("cskey")
	if len(cskey) == 0 {
		default_error_response(w, r, fmt.Errorf("contextual_signal_key not present"))
	}

	csvalue := r.FormValue("csvalue")
	if len(csvalue) == 0 {
		default_error_response(w, r, fmt.Errorf("contextual_signal_value not present"))
	}

	ig := r.FormValue("ig")
	if len(ig) == 0 {
		default_error_response(w, r, fmt.Errorf("interest_group not present"))
	}

	liid := GetInt(r.FormValue("liid"))
	if _, ok := lineItemMap[liid]; !ok {
		default_error_response(w, r, fmt.Errorf("lineitem not present : %d", liid))
		return
	}

	AddCSIGLIMapping(CSIGKey(cskey, csvalue, ig), liid)

	success_response(w, r, SUCCESS)
}
*/
func kv_get_csig(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := r.FormValue("key")
	if len(key) == 0 {
		default_error_response(w, r, fmt.Errorf("key not present"))
	}

	result := GetResult(key)
	response, _ := json.Marshal(result)
	success_response(w, r, string(response))
}

//impression_count
func kv_update_ic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	/*
		[
		{
			"auc_lineitem_id": 101,
			"imp_count": 1,
			"winning_imp_count": 1
		},
		{
			"auc_lineitem_id": 102,
			"imp_count": 2,
			"winning_imp_count": 2
		}
		]
	*/
	logger.Debug("Update LineItem Analytics Numbers")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		default_error_response(w, r, err)
	}
	r.Body.Close()

	type temp struct {
		LineItemID   int `json:"auc_lineitem_id,omitempty"`
		ImpCount     int `json:"imp_count,omitempty"`
		WinningCount int `json:"winning_imp_count,omitempty"`
	}
	result := []temp{}
	if err := json.Unmarshal(body, result); nil != err {
		default_error_response(w, r, err)
	}

	for _, item := range result {
		if li, ok := lineItemMap[item.LineItemID]; ok {
			li.updateImpressions(item.ImpCount, item.WinningCount)
		}
	}

	success_response(w, r, SUCCESS)
}

//impression_count
func flush(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	FlushAll()
	success_response(w, r, SUCCESS)
}
