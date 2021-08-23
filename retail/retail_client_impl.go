package retail

import (
	"errors"
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	. "github.com/byteplus-sdk/sdk-go/retail/protocol"
	"strings"
)

var (
	writeMsgFormat  = "Only can receive max to %d items in one write request"
	writeTooManyErr = errors.New(fmt.Sprintf(writeMsgFormat, MaxWriteItemCount))

	importMsgFormat  = "Only can receive max to %d items in one import request"
	importTooManyErr = errors.New(fmt.Sprintf(importMsgFormat, MaxImportItemCount))
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
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteUsers] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ImportUsers(request *ImportUsersRequest,
	opts ...option.Option) (*OperationResponse, error) {
	users := request.GetInputConfig().GetUsersInlineSource().GetUsers()
	if len(users) > MaxImportItemCount {
		return nil, importTooManyErr
	}
	url := c.ru.importUsersURL
	response := &OperationResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[ImportUsers] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) WriteProducts(request *WriteProductsRequest,
	opts ...option.Option) (*WriteProductsResponse, error) {
	if len(request.Products) > MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.ru.writeProductsURL
	response := &WriteProductsResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteProducts] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ImportProducts(request *ImportProductsRequest,
	opts ...option.Option) (*OperationResponse, error) {
	products := request.GetInputConfig().GetProductsInlineSource().GetProducts()
	if len(products) > MaxImportItemCount {
		return nil, importTooManyErr
	}
	url := c.ru.importProductsURL
	response := &OperationResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[ImportProducts] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) WriteUserEvents(request *WriteUserEventsRequest,
	opts ...option.Option) (*WriteUserEventsResponse, error) {
	if len(request.UserEvents) > MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.ru.writeUserEventsURL
	response := &WriteUserEventsResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteUserEvents] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ImportUserEvents(request *ImportUserEventsRequest,
	opts ...option.Option) (*OperationResponse, error) {
	userEvents := request.GetInputConfig().GetUserEventsInlineSource().GetUserEvents()
	if len(userEvents) > MaxImportItemCount {
		return nil, importTooManyErr
	}
	url := c.ru.importUserEventsURL
	response := &OperationResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[ImportUserEvents] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Predict(request *PredictRequest, scene string,
	opts ...option.Option) (*PredictResponse, error) {
	url := strings.ReplaceAll(c.ru.predictURLFormat, "{}", scene)
	response := &PredictResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
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
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
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
