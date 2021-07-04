package general

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
)

const (
	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/home
	predictUrlFormat = "%s://%s/predict/api/%s/{}"

	// The URL format of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/callback
	callbackUrlFormat = "%s://%s/predict/api/%s/callback"

	// The URL format of data uploading
	// Example: https://tob.sgsnssdk.com/data/api/general_demo/user?method=write
	uploadUrlFormat = "%s://%s/data/api/%s/{}?method=%s"

	// The URL format of marking a whole day data has been imported completely
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/done?topic=user
	doneUrlFormat = "%s://%s/data/api/%s/done?topic={}"
)

type generalURL struct {
	cu     *common.URL
	schema string
	tenant string

	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/home
	predictUrlFormat string

	// The URL of reporting the real exposure list
	// Example: https://tob.sgsnssdk.com/predict/api/general_demo/callback
	callbackURL string

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

func (receiver *generalURL) Refresh(host string) {
	receiver.cu.Refresh(host)
	receiver.predictUrlFormat = receiver.generatePredictURLFormat(host)
	receiver.callbackURL = receiver.generateCallbackURL(host)
	receiver.writeDataURLFormat = receiver.generateUploadURL(host, "write")
	receiver.importDataURLFormat = receiver.generateUploadURL(host, "import")
	receiver.doneURLFormat = receiver.generateDoneURL(host)
}

func (receiver *generalURL) generatePredictURLFormat(host string) string {
	return fmt.Sprintf(predictUrlFormat, receiver.schema, host, receiver.tenant)
}

func (receiver *generalURL) generateCallbackURL(host string) string {
	return fmt.Sprintf(callbackUrlFormat, receiver.schema, host, receiver.tenant)
}

func (receiver *generalURL) generateUploadURL(host string, method string) string {
	return fmt.Sprintf(uploadUrlFormat, receiver.schema, host, receiver.tenant, method)
}

func (receiver *generalURL) generateDoneURL(host string) string {
	return fmt.Sprintf(doneUrlFormat, receiver.schema, host, receiver.tenant)
}
