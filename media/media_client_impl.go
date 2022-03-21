package media

import (
	"errors"
	"fmt"

	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	"github.com/byteplus-sdk/sdk-go/media/protocol"
)

var (
	writeMsgFormat  = "Only can receive max to %d items in one write request"
	writeTooManyErr = errors.New(fmt.Sprintf(writeMsgFormat, core.MaxWriteItemCount))
)

type clientImpl struct {
	common.Client
	hCaller *core.HTTPCaller
	mu      *mediaURL
	hostAva *core.HostAvailabler
}

func (c clientImpl) WriteUsers(request *protocol.WriteUsersRequest,
	opts ...option.Option) (*protocol.WriteUsersResponse, error) {
	if len(request.Users) > core.MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.mu.writeUsersURL
	response := &protocol.WriteUsersResponse{}
	err := c.hCaller.DoPBRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteUsers] rsp:\n%s\n", response)
	return response, nil
}

func (c clientImpl) WriteContents(request *protocol.WriteContentsRequest,
	opts ...option.Option) (*protocol.WriteContentsResponse, error) {
	if len(request.Contents) > core.MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.mu.writeContentsURL
	response := &protocol.WriteContentsResponse{}
	err := c.hCaller.DoPBRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteContents] rsp:\n%s\n", response)
	return response, nil
}

func (c clientImpl) WriteUserEvents(request *protocol.WriteUserEventsRequest,
	opts ...option.Option) (*protocol.WriteUserEventsResponse, error) {
	if len(request.UserEvents) > core.MaxWriteItemCount {
		return nil, writeTooManyErr
	}
	url := c.mu.writeUserEventsURL
	response := &protocol.WriteUserEventsResponse{}
	err := c.hCaller.DoPBRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteUserEvents] rsp:\n%s\n", response)
	return response, nil
}

func (c clientImpl) Release() {
	c.hostAva.Shutdown()
}
