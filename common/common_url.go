package common

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core"
)

const (
	// The URL format of operation information
	// Example: https://tob.sgsnssdk.com/data/api/retail_demo/operation?method=get
	operationUrlFormat = "%s://%s/data/api/%s/operation?method=%s"
)

func NewURL(context *core.Context) *URL {
	return &URL{
		schema: context.Schema(),
		tenant: context.Tenant(),
	}
}

type URL struct {
	schema            string
	tenant            string
	getOperationUrl   string
	listOperationsUrl string
}

func (receiver *URL) Refresh(host string) {
	receiver.getOperationUrl = receiver.generateOperationUrl(host, "get")
	receiver.listOperationsUrl = receiver.generateOperationUrl(host, "list")
}

func (receiver *URL) generateOperationUrl(host string, method string) string {
	return fmt.Sprintf(operationUrlFormat, receiver.schema, host, receiver.tenant, method)
}
