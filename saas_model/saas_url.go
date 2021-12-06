package saas_model

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
)

const (
	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/home
	predictURLFormat = "%s://%s/saasmodel/%s/model/{}"

	// The URL format of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/ack_impression
	ackImpressionURLFormat = "%s://%s/saasmodel/%s/model/{}/ack_server_impressions"

	// The URL format of data uploading
	// Example: https://tob.sgsnssdk.com/data/api/general_demo/user?method=write
	uploadUrlFormat = "%s://%s/data/api/%s/{}?method=%s"

	// The URL format of marking a whole day data has been imported completely
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/done?topic=user
	doneUrlFormat = "%s://%s/data/api/%s/done?topic={}"
)

type saasURL struct {
	su        *common.URL
	schema    string
	projectId string

	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/home
	predictURLFormat string

	// The URL of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/retail/demo/ack_server_impression
	ackImpressionURLFormat string

	// The URL of uploading real-time user data
	// Example: https://tob.sgsnssdk.com/data/api/general_demo/user?method=write
	writeDataURLFormat string

	// The URL of importing daily offline user data
	// Example: https://tob.sgsnssdk.com/data/api/general_demo/user?method=import
	importDataURLFormat string

	// The URL format of marking a whole day data has been imported completely
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/done?topic=user
	doneURLFormat string
}

func (receiver *saasURL) Refresh(host string) {
	receiver.su.Refresh(host)
	receiver.predictURLFormat = receiver.generatePredictURLFormat(host)
	receiver.ackImpressionURLFormat = receiver.generateAckURL(host)
	receiver.writeDataURLFormat = receiver.generateUploadURL(host, "write")
	receiver.importDataURLFormat = receiver.generateUploadURL(host, "import")
	receiver.doneURLFormat = receiver.generateDoneURL(host)
}

func (receiver *saasURL) generatePredictURLFormat(host string) string {
	return fmt.Sprintf(predictURLFormat, receiver.schema, host, receiver.projectId)
}

func (receiver *saasURL) generateAckURL(host string) string {
	return fmt.Sprintf(ackImpressionURLFormat, receiver.schema, host, receiver.projectId)
}

func (receiver *saasURL) generateUploadURL(host string, method string) string {
	return fmt.Sprintf(uploadUrlFormat, receiver.schema, host, receiver.projectId, method)
}

func (receiver *saasURL) generateDoneURL(host string) string {
	return fmt.Sprintf(doneUrlFormat, receiver.schema, host, receiver.projectId)
}
