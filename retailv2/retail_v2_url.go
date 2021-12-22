package retailv2

import (
	"fmt"

	"github.com/byteplus-sdk/sdk-go/common"
)

const (
	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/home
	predictURLFormat = "%s://%s/predict/api/retail/%s/{}"

	// The URL format of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/ack_impression
	ackImpressionURLFormat = "%s://%s/predict/api/retail/%s/ack_server_impressions"

	// The URL format of data uploading
	// Example: https://tob.sgsnssdk.com/data/api/retail/retail_demo/user?method=write
	uploadURLFormat = "%s://%s/data/api/retail/v2/%s/%s?method=%s"
)

type retailURL struct {
	cu     *common.URL
	schema string
	tenant string

	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/home
	predictURLFormat string

	// The URL of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/ack_server_impression
	ackImpressionURL string

	// The URL of uploading real-time user data
	// Example: https://tob.sgsnssdk.com/data/api/retail/retail_demo/user?method=write
	writeUsersURL string

	// The URL of uploading real-time product data
	// Example: https://tob.sgsnssdk.com/data/api/retail/retail_demo/product?method=write
	writeProductsURL string

	// The URL of uploading real-time user event data
	// Example: https://tob.sgsnssdk.com/data/api/retail/retail_demo/user_event?method=write
	writeUserEventsURL string
}

func (receiver *retailURL) Refresh(host string) {
	receiver.cu.Refresh(host)
	receiver.predictURLFormat = receiver.generatePredictURLFormat(host)
	receiver.ackImpressionURL = receiver.generateAckURL(host)
	receiver.writeUsersURL = receiver.generateUploadURL(host, "user", "write")
	receiver.writeProductsURL = receiver.generateUploadURL(host, "product", "write")
	receiver.writeUserEventsURL = receiver.generateUploadURL(host, "user_event", "write")
}

func (receiver *retailURL) generatePredictURLFormat(host string) string {
	return fmt.Sprintf(predictURLFormat, receiver.schema, host, receiver.tenant)
}

func (receiver *retailURL) generateAckURL(host string) string {
	return fmt.Sprintf(ackImpressionURLFormat, receiver.schema, host, receiver.tenant)
}

func (receiver *retailURL) generateUploadURL(host string, topic string, method string) string {
	return fmt.Sprintf(uploadURLFormat, receiver.schema, host, receiver.tenant, topic, method)
}
