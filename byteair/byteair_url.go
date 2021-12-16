package byteair

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
)

const (
	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://byteair-api-cn1.snssdk.com/predict/api/20013144/home
	predictUrlFormat = "%s://%s/predict/api/%s/{}"

	// The URL format of reporting the real exposure list
	// Example: https://byteair-api-cn1.snssdk.com/predict/api/20013144/callback
	callbackUrlFormat = "%s://%s/predict/api/%s/callback"

	// The URL format of data uploading
	// Example: https://byteair-api-cn1.snssdk.com/data/api/20013144/user?method=write
	uploadUrlFormat = "%s://%s/data/api/%s/{}?method=%s"

	// The URL format of marking a whole day data has been imported completely
	// Example: https://byteair-api-cn1.snssdk.com/data/api/20013144/done?topic=user
	doneUrlFormat = "%s://%s/data/api/%s/done?topic={}"
)

type byteairURL struct {
	cu     *common.URL
	schema string
	tenant string

	// The URL template of "predict" request, which need fill with "scene" info when use
	// Example: https://byteair-api-cn1.snssdk.com/predict/api/20013144/home
	predictUrlFormat string

	// The URL of reporting the real exposure list
	// Example: https://byteair-api-cn1.snssdk.com/predict/api/20013144/callback
	callbackURL string

	// The URL of uploading real-time user data
	// Example: https://byteair-api-cn1.snssdk.com/data/api/20013144/user?method=write
	writeDataURLFormat string

	// The URL of importing daily offline user data
	// Example: https://byteair-api-cn1.snssdk.com/data/api/20013144/user?method=import
	importDataURLFormat string

	// The URL format of marking a whole day data has been imported completely
	// Example: https://byteair-api-cn1.snssdk.com/data/api/20013144/done?topic=user
	doneURLFormat string
}

func (receiver *byteairURL) Refresh(host string) {
	receiver.cu.Refresh(host)
	receiver.predictUrlFormat = receiver.generatePredictURLFormat(host)
	receiver.callbackURL = receiver.generateCallbackURL(host)
	receiver.writeDataURLFormat = receiver.generateUploadURL(host, "write")
	receiver.importDataURLFormat = receiver.generateUploadURL(host, "import")
	receiver.doneURLFormat = receiver.generateDoneURL(host)
}

func (receiver *byteairURL) generatePredictURLFormat(host string) string {
	return fmt.Sprintf(predictUrlFormat, receiver.schema, host, receiver.tenant)
}

func (receiver *byteairURL) generateCallbackURL(host string) string {
	return fmt.Sprintf(callbackUrlFormat, receiver.schema, host, receiver.tenant)
}

func (receiver *byteairURL) generateUploadURL(host string, method string) string {
	return fmt.Sprintf(uploadUrlFormat, receiver.schema, host, receiver.tenant, method)
}

func (receiver *byteairURL) generateDoneURL(host string) string {
	return fmt.Sprintf(doneUrlFormat, receiver.schema, host, receiver.tenant)
}
