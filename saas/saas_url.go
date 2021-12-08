package saas

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
)

const (
	// The URL template of "predict" request
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/Predict
	predictURLFormat = "%s://%s/RetailSaaS/Predict"

	// The URL format of reporting the real exposure list
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/AckServerImpressions
	ackImpressionURLFormat = "%s://%s/RetailSaaS/AckServerImpressions"

	// The URL format of data uploading
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/WriteUsers
	uploadUrlFormat = "%s://%s/RetailSaaS/%s"
)

type saasURL struct {
	su        *common.URL
	schema    string
	projectId string

	// The URL template of "predict" request
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/Predict
	predictURL string

	// The URL of reporting the real exposure list
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/AckServerImpressions
	ackImpressionURL string

	// The URL of uploading real-time user data
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/WriteUsers
	writeUsersDataURL string

	// The URL of uploading real-time product data
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/WriteProducts
	writeProductsDataURL string

	// The URL of uploading real-time user event data
	// Example: https://rec-api-sg1.recplusapi.com/RetailSaaS/WriteUserEvents
	writeUserEventsDataURL string
}

func (receiver *saasURL) Refresh(host string) {
	receiver.su.Refresh(host)
	receiver.predictURL = receiver.generatePredictURLFormat(host)
	receiver.ackImpressionURL = receiver.generateAckURL(host)
	receiver.writeUsersDataURL = receiver.generateUploadURL(host, "WriteUsers")
	receiver.writeProductsDataURL = receiver.generateUploadURL(host, "WriteProducts")
	receiver.writeUserEventsDataURL = receiver.generateUploadURL(host, "WriteUserEvents")
}

func (receiver *saasURL) generatePredictURLFormat(host string) string {
	return fmt.Sprintf(predictURLFormat, receiver.schema, host)
}

func (receiver *saasURL) generateAckURL(host string) string {
	return fmt.Sprintf(ackImpressionURLFormat, receiver.schema, host)
}

func (receiver *saasURL) generateUploadURL(host string, topic string) string {
	return fmt.Sprintf(uploadUrlFormat, receiver.schema, host, topic)
}
