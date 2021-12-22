package common

import (
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	"strings"
	"time"
)

func NewClient(cli *HttpCaller, cu *URL) Client {
	return &clientImpl{
		cli: cli,
		cu:  cu,
	}
}

type clientImpl struct {
	cli *HttpCaller
	cu  *URL
}

func (c *clientImpl) GetOperation(request *GetOperationRequest,
	opts ...option.Option) (*OperationResponse, error) {
	url := c.cu.getOperationUrl
	response := &OperationResponse{}
	err := c.cli.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[GetOperations] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ListOperations(request *ListOperationsRequest,
	opts ...option.Option) (*ListOperationsResponse, error) {
	url := c.cu.listOperationsUrl
	response := &ListOperationsResponse{}
	err := c.cli.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[ListOperation] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Done(dateList []time.Time, topic string, opts ...option.Option) (*DoneResponse, error) {
	var dates []*Date
	if len(dateList) == 0 {
		previousDay := time.Now().Add(-24 * time.Hour)
		dates = c.appendDoneDate(dates, previousDay)
	} else {
		for _, date := range dateList {
			dates = c.appendDoneDate(dates, date)
		}
	}
	url := strings.ReplaceAll(c.cu.doneUrlFormat, "{}", topic)
	request := &DoneRequest{
		DataDates: dates,
	}
	response := &DoneResponse{}
	err := c.cli.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[Done] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) appendDoneDate(dates []*Date,
	date time.Time) []*Date {
	return append(dates, &Date{
		Year:  int32(date.Year()),
		Month: int32(date.Month()),
		Day:   int32(date.Day()),
	})
}
