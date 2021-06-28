package retail

import (
	"errors"
	"fmt"
	. "github.com/byte-plus/sdk-go/core"
	"github.com/byte-plus/sdk-go/core/logs"
	"github.com/byte-plus/sdk-go/core/option"
	. "github.com/byte-plus/sdk-go/retail/protocol"
	"strings"
)

var (
	writeMsgFormat  = "Only can receive %d items in one write request"
	writeTooManyErr = errors.New(fmt.Sprintf(writeMsgFormat, MaxWriteItemCount))

	importMsgFormat  = "Only can receive %d items in one import request"
	importTooManyErr = errors.New(fmt.Sprintf(importMsgFormat, MaxImportItemCount))
)

type clientImpl struct {
	httpCli *HttpCaller
	ru      *retailURL
}

func (c *clientImpl) WriteUsers(request *WriteUsersRequest,
	opts ...option.Option) (*WriteUsersResponse, error) {
	if len(request.Users) > MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.ru.writeUsersUrl
	response := &WriteUsersResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
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
	url := c.ru.importUsersUrl
	response := &OperationResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
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
	url := c.ru.writeProductsUrl
	response := &WriteProductsResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
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
	url := c.ru.importProductsUrl
	response := &OperationResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
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
	url := c.ru.writeUserEventsUrl
	response := &WriteUserEventsResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
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
	url := c.ru.importUserEventsUrl
	response := &OperationResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[ImportUserEvents] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) GetOperation(request *GetOperationRequest,
	opts ...option.Option) (*OperationResponse, error) {
	url := c.ru.getOperationUrl
	response := &OperationResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[GetOperations] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ListOperations(request *ListOperationsRequest,
	opts ...option.Option) (*ListOperationsResponse, error) {
	url := c.ru.listOperationsUrl
	response := &ListOperationsResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[ListOperation] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Predict(request *PredictRequest, scene string,
	opts ...option.Option) (*PredictResponse, error) {
	url := strings.ReplaceAll(c.ru.predictUrlFormat, "{}", scene)
	response := &PredictResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[Predict] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) AckServerImpressions(request *AckServerImpressionsRequest,
	opts ...option.Option) (*AckServerImpressionsResponse, error) {
	url := c.ru.ackImpressionUrl
	response := &AckServerImpressionsResponse{}
	err := c.httpCli.DoRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[AckImpressions] rsp:\n%s\n", response)
	return response, nil
}
