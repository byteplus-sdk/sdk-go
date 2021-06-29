package retail

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core"
)

const (
	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/home
	predictUrlFormat = "%s://%s/predict/api/retail/%s/{}"

	// The URL format of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/ack_impression
	ackImpressionUrlFormat = "%s://%s/predict/api/retail/%s/ack_server_impressions"

	// The URL format of data uploading
	// Example: https://tob.sgsnssdk.com/data/api/retail/retail_demo/user?method=write
	uploadUrlFormat = "%s://%s/data/api/retail/%s/%s?method=%s"

	// The URL format of operation information
	// Example: https://tob.sgsnssdk.com/data/api/retail/retail_demo/operation?method=get
	operationUrlFormat = "%s://%s/data/api/retail/%s/operation?method=%s"
)

type retailURL struct {
	context             *core.Context
	hostAvailabler      *core.HostAvailabler
	predictUrlFormat    string
	ackImpressionUrl    string
	writeUsersUrl       string
	importUsersUrl      string
	writeProductsUrl    string
	importProductsUrl   string
	writeUserEventsUrl  string
	importUserEventsUrl string
	getOperationUrl     string
	listOperationsUrl   string
}

func (receiver *retailURL) Refresh(host string) {
	receiver.predictUrlFormat = receiver.generatePredictUrl(host)
	receiver.ackImpressionUrl = receiver.generateAckUrl(host)
	receiver.writeUsersUrl = receiver.generateUploadUrl(host, "user", "write")
	receiver.importUsersUrl = receiver.generateUploadUrl(host, "user", "import")
	receiver.writeProductsUrl = receiver.generateUploadUrl(host, "product", "write")
	receiver.importProductsUrl = receiver.generateUploadUrl(host, "product", "import")
	receiver.writeUserEventsUrl = receiver.generateUploadUrl(host, "user_event", "write")
	receiver.importUserEventsUrl = receiver.generateUploadUrl(host, "user_event", "import")
	receiver.getOperationUrl = receiver.generateOperationUrl(host, "get")
	receiver.listOperationsUrl = receiver.generateOperationUrl(host, "list")
}

func (receiver *retailURL) generatePredictUrl(host string) string {
	schema := receiver.context.Schema()
	tenant := receiver.context.Tenant()
	return fmt.Sprintf(predictUrlFormat, schema, host, tenant)
}

func (receiver *retailURL) generateAckUrl(host string) string {
	schema := receiver.context.Schema()
	tenant := receiver.context.Tenant()
	return fmt.Sprintf(ackImpressionUrlFormat, schema, host, tenant)
}

func (receiver *retailURL) generateUploadUrl(host string, topic string, method string) string {
	schema := receiver.context.Schema()
	tenant := receiver.context.Tenant()
	return fmt.Sprintf(uploadUrlFormat, schema, host, tenant, topic, method)
}

func (receiver *retailURL) generateOperationUrl(host string, method string) string {
	schema := receiver.context.Schema()
	tenant := receiver.context.Tenant()
	return fmt.Sprintf(operationUrlFormat, schema, host, tenant, method)
}
