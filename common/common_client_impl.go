package common

import (
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
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
