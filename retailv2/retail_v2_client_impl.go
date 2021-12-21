package retailv2

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/byteplus-sdk/sdk-go/common"
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	. "github.com/byteplus-sdk/sdk-go/retailv2/protocol"
)

var (
	writeMsgFormat  = "Only can receive max to %d items in one write request"
	writeTooManyErr = errors.New(fmt.Sprintf(writeMsgFormat, MaxWriteItemCount))
)

type clientImpl struct {
	cCli    common.Client
	hCaller *HttpCaller
	ru      *retailURL
	hostAva *HostAvailabler
}

func (c *clientImpl) Release() {
	c.hostAva.Shutdown()
}

func (c *clientImpl) WriteUsers(request *WriteUsersRequest,
	opts ...option.Option) (*WriteUsersResponse, error) {
	if len(request.Users) > MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.ru.writeUsersURL
	response := &WriteUsersResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteUsers] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) WriteProducts(request *WriteProductsRequest,
	opts ...option.Option) (*WriteProductsResponse, error) {
	if len(request.Products) > MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.ru.writeProductsURL
	response := &WriteProductsResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteProducts] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) WriteUserEvents(request *WriteUserEventsRequest,
	opts ...option.Option) (*WriteUserEventsResponse, error) {
	if len(request.UserEvents) > MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.ru.writeUserEventsURL
	response := &WriteUserEventsResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteUserEvents] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Predict(request *PredictRequest, scene string,
	opts ...option.Option) (*PredictResponse, error) {
	url := strings.ReplaceAll(c.ru.predictURLFormat, "{}", scene)
	response := &PredictResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[Predict] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) AckServerImpressions(request *AckServerImpressionsRequest,
	opts ...option.Option) (*AckServerImpressionsResponse, error) {
	url := c.ru.ackImpressionURL
	response := &AckServerImpressionsResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[AckImpressions] rsp:\n%s\n", response)
	return response, nil
}

// GetOperation
//
// Gets the operation of a previous long running call.
func (c *clientImpl) GetOperation(request *GetOperationRequest,
	opts ...option.Option) (*OperationResponse, error) {
	return c.cCli.GetOperation(request, opts...)
}

// ListOperations
//
// Lists operations that match the specified filter in the request.
func (c *clientImpl) ListOperations(request *ListOperationsRequest,
	opts ...option.Option) (*ListOperationsResponse, error) {
	return c.cCli.ListOperations(request, opts...)
}

// Done
// Pass a date list to mark the completion of data synchronization for these days
// suitable for new API
func (c *clientImpl) Done(dateList []time.Time, topic string, opts ...option.Option) (*Response, error) {
	var dates []*Date
	if len(dateList) == 0 {
		previousDay := time.Now().Add(-24 * time.Hour)
		dates = c.appendDoneDate(dates, previousDay)
	} else {
		for _, date := range dateList {
			dates = c.appendDoneDate(dates, date)
		}
	}
	url := strings.ReplaceAll(c.ru.doneUrlFormat, "{}", topic)
	request := &DoneRequest{
		DataDates: dates,
	}
	response := &Response{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
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
