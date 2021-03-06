/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 9/9/20, 9:41 AM
 */

package action

import (
	"attribution/pkg/logic"
	"attribution/pkg/handler/http-handler/click/data"
	"attribution/pkg/storage"

	"github.com/golang/glog"
)

type ClickLogIndexAction struct {
	index storage.ClickIndex
}

func NewClickLogIndexAction(index storage.ClickIndex) *ClickLogIndexAction {
	return &ClickLogIndexAction{
		index: index,
	}
}

func (action *ClickLogIndexAction) Run(i interface{}) {
	c := i.(*data.ClickContext)

	if err := action.runInternal(c); err != nil {
		c.StopWithError(err)
		glog.Errorf("err: %v", err)
	}
}

func (action *ClickLogIndexAction) runInternal(c *data.ClickContext) error {
	return logic.ProcessClickLog(c.ClickLog, action.index)
}
