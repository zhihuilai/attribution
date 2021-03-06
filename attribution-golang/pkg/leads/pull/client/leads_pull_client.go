/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 9/25/20, 2:34 PM
 */

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"attribution/pkg/leads/pull/protocal"
	"attribution/pkg/oauth"
	"attribution/pkg/storage"

	"github.com/golang/glog"
)

// 负责从线索平台拉取线索信息
type LeadsPullClient struct {
	config     *PullConfig
	httpClient *http.Client
	storage    storage.LeadsStorage

	// 进度
	beginTime    time.Time     // 拉取开始的时间
	endTime      time.Time     // 拉取结束的时间戳
	pullInterval time.Duration // 拉取间隔

	nextPage     int // 下次拉取第几页
	totalPage    int // 总共有多少页
	currentCount int // 当前时间段总共的线索数

	lastSearch [2]string // 最后查到的线索，用户深度翻页
}

type PullConfig struct {
	AccountId int64                    `json:"account_id"`
	Filtering []map[string]interface{} `json:"filtering"`
	PageSize  int                      `json:"page_size"`
}

func NewLeadsPullClient(config *PullConfig) *LeadsPullClient {
	return &LeadsPullClient{
		config:     config,
		httpClient: &http.Client{},
	}
}
func (c *LeadsPullClient) WithStorage(storage storage.LeadsStorage) *LeadsPullClient {
	c.storage = storage
	return c
}

func (c *LeadsPullClient) formatRequestUrl() (string, error) {
	token, err := oauth.GetToken()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`https://api.e.qq.com/v1.3/lead_clues/get?access_token=%s&timestamp=%d&nonce=%s`,
		token, time.Now().Unix(), oauth.GenNonce()), nil
}

func (c *LeadsPullClient) Pull() error {

}

func (c *LeadsPullClient) pullRoutine() error {
	for {
		if time.Now().Sub(c.beginTime) >= c.pullInterval {
			if err := c.requestInterval(); err != nil {
				glog.Errorf("failed to pull interval, begin: %v, end: %v", c.beginTime, c.endTime)
			} else {
				c.beginTime = c.endTime
				c.endTime = c.endTime.Add(c.pullInterval)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (c *LeadsPullClient) requestInterval() error {

}

func (c *LeadsPullClient) requestPage() error {
	url, err := c.formatRequestUrl()
	if err != nil {
		return err
	}

	config := c.config
	req := &protocal.LeadsRequest{
		AccountId: config.AccountId,
		TimeRange: &protocal.TimeRange{
			StartTime: c.beginTime.Unix(),
			EndTime:   c.endTime.Unix(),
		},
		Filtering: config.Filtering,
		PageSize:  config.PageSize,
	}

	if c.currentCount+config.PageSize <= 5000 {
		req.Page = c.nextPage
	} else {
		req.LastSearchAfterValues = c.lastSearch
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		glog.Errorf("failed to create request, err: %v", err)
		return err
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	if httpResp.StatusCode != 200 {
		return fmt.Errorf("http response status[%s] not valid", httpResp.Status)
	}
	respBody, err := ioutil.ReadAll(httpResp.Body)

	var resp protocal.LeadsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("leads response code[%d] not valid, msg: %s", resp.Code, resp.Message)
	}
	if len(resp.Data) == 0 {
		return errors.New("leads data empty")
	}
	if err := c.store(&resp); err != nil {
		return err
	}
	c.onPageRequestSuccess(&resp)
	return nil
}

func (c *LeadsPullClient) store(resp *protocal.LeadsResponse) error {
	for _, info := range resp.Data {
		if err := c.storage.Store(info); err != nil {
			return err
		}
	}
	return nil
}

func (c *LeadsPullClient) onPageRequestSuccess(resp *protocal.LeadsResponse) {
	last := resp.Data[len(resp.Data)-1]
	c.lastSearch = [2]string{
		strconv.FormatInt(last.LeadsActionTime, 64),
		strconv.FormatInt(last.LeadsId, 64),
	}

	c.nextPage++
}
