package common

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core"
)

const (
	// The URL format of operation information
	// Example: https://tob.sgsnssdk.com/data/api/retail_demo/operation?method=get
	operationUrlFormat = "%s://%s/data/api/%s/operation?method=%s"

	// The URL of mark certain days that data synchronization is complete
	// Example: https://tob.sgsnssdk.com/data/api/retail_demo/done?topic=user
	doneUrlFormat = "%s://%s/data/api/%s/done?topic={}"
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
	doneUrlFormat     string
}

func (receiver *URL) Refresh(host string) {
	receiver.getOperationUrl = receiver.generateOperationUrl(host, "get")
	receiver.listOperationsUrl = receiver.generateOperationUrl(host, "list")
	receiver.doneUrlFormat = receiver.generateDoneUrl(host)
}

func (receiver *URL) generateOperationUrl(host string, method string) string {
	return fmt.Sprintf(operationUrlFormat, receiver.schema, host, receiver.tenant, method)
}

func (receiver *URL) generateDoneUrl(host string) string {
	return fmt.Sprintf(doneUrlFormat, receiver.schema, host, receiver.tenant)
}
