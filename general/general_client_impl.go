package general

import (
	"errors"
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	. "github.com/byteplus-sdk/sdk-go/general/protocol"
	"strings"
)

var (
	errMsgFormat    = "Only can receive max to %d items in one request"
	TooManyItemsErr = errors.New(fmt.Sprintf(errMsgFormat, MaxImportItemCount))
)

type clientImpl struct {
	common.Client
	hCaller *HttpCaller
	gu      *generalURL
	hostAva *HostAvailabler
}

func (c *clientImpl) Release() {
	c.hostAva.Shutdown()
}

func (c *clientImpl) WriteData(dataList []map[string]interface{}, topic string,
	opts ...option.Option) (*WriteResponse, error) {
	if len(dataList) > MaxImportItemCount {
		return nil, TooManyItemsErr
	}
	urlFormat := c.gu.writeDataURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	response := &WriteResponse{}
	err := c.hCaller.DoJsonRequest(url, dataList, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteData] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Predict(request *PredictRequest,
	scene string, opts ...option.Option) (*PredictResponse, error) {
	urlFormat := c.gu.predictUrlFormat
	url := strings.ReplaceAll(urlFormat, "{}", scene)
	response := &PredictResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[Predict] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Callback(request *CallbackRequest,
	opts ...option.Option) (*CallbackResponse, error) {
	url := c.gu.callbackURL
	response := &CallbackResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[Callback] rsp:\n%s\n", response)
	return response, nil
}
