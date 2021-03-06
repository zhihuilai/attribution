/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 9/25/20, 5:38 PM
 */

package handle

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"attribution/pkg/leads/pull/protocal"
	"attribution/pkg/storage"

	"github.com/golang/glog"
)

type ReceiveLeadsHandle struct {
	leadsStorage storage.LeadsStorage
}

func NewReceiveLeadsHandle() *ReceiveLeadsHandle {
	return &ReceiveLeadsHandle{}
}

func (handle *ReceiveLeadsHandle) WithLeadsStorage(storage storage.LeadsStorage) *ReceiveLeadsHandle {
	handle.leadsStorage = storage
	return handle
}

type Resp struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func (handle *ReceiveLeadsHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *Resp
	if err := handle.doServeHttp(w, r); err != nil {
		glog.Errorf("failed to serve, err: %v", err)
		// TODO(监控)

		resp = &Resp{
			Code: -1,
			Message: err.Error(),
		}
	} else {
		resp = &Resp{
			Code: 0,
			Message: "success",
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	bytes, err := json.Marshal(resp)
	if err == nil {
		if _, writeErr := w.Write(bytes); writeErr != nil {
			glog.Errorf("failed to write data, err: %v", writeErr)
		}
	} else {
		glog.Errorf("failed to marshal response, err: %v", err)
	}
}

func (handle *ReceiveLeadsHandle) doServeHttp(_ http.ResponseWriter, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var leads protocal.LeadsInfo
	if err := json.Unmarshal(body, &leads); err != nil {
		return err
	}

	if err := handle.leadsStorage.Store(&leads); err != nil {
		return err
	}

	return nil
}

