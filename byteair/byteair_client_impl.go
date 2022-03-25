package byteair

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/byteplus-sdk/sdk-go/byteair/protocol"
	"github.com/byteplus-sdk/sdk-go/common"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
)

const (
	DefaultPredictScene  = "default"
	DefaultCallbackScene = "default"
)

var (
	errMsgFormat    = "Only can receive max to %d items in one request"
	TooManyItemsErr = errors.New(fmt.Sprintf(errMsgFormat, MaxImportItemCount))
)

type clientImpl struct {
	common.Client
	hCaller *HTTPCaller
	gu      *byteairURL
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
	err := c.hCaller.DoJSONRequest(url, dataList, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteData] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Predict(request *PredictRequest,
	opts ...option.Option) (*PredictResponse, error) {
	urlFormat := c.gu.predictUrlFormat
	//The options conversion should be placed in xxx_client_impl,
	//so that each client_impl could do some special processing according to options
	options := option.Conv2Options(opts...)
	scene := options.Scene
	// If predict scene option is not filled, add default value
	if scene == "" {
		scene = DefaultPredictScene
	}
	url := strings.ReplaceAll(urlFormat, "{}", scene)
	response := &PredictResponse{}
	err := c.hCaller.DoPBRequest(url, request, response, options)
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
	// If predict scene option is not filled, add default value
	if request.Scene == "" {
		request.Scene = DefaultCallbackScene
	}
	err := c.hCaller.DoPBRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[Callback] rsp:\n%s\n", response)
	return response, nil
}
