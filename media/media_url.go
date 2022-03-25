package media

import (
	"fmt"

	"github.com/byteplus-sdk/sdk-go/common"
)

const (
	// The URL format of data uploading
	// Example: https://tob.sgsnssdk.com/data/api/media/media_demo/user?method=write
	uploadURLFormat = "%s://%s/data/api/media/%s/%s?method=%s"
)

type mediaURL struct {
	cu     *common.URL
	schema string
	tenant string

	// The URL of uploading real-time user data
	// Example: https://tob.sgsnssdk.com/data/api/media/media_demo/user?method=write
	writeUsersURL string

	// The URL of uploading real-time content data
	// Example: https://tob.sgsnssdk.com/data/api/media/media_demo/content?method=write
	writeContentsURL string

	// The URL of uploading real-time user event data
	// Example: https://tob.sgsnssdk.com/data/api/media/media_demo/user_event?method=write
	writeUserEventsURL string
}

func (receiver *mediaURL) Refresh(host string) {
	receiver.cu.Refresh(host)
	receiver.writeUsersURL = receiver.generateUploadURL(host, "user", "write")
	receiver.writeContentsURL = receiver.generateUploadURL(host, "content", "write")
	receiver.writeUserEventsURL = receiver.generateUploadURL(host, "user_event", "write")
}

func (receiver *mediaURL) generateUploadURL(host string, topic string, method string) string {
	return fmt.Sprintf(uploadURLFormat, receiver.schema, host, receiver.tenant, topic, method)
}
