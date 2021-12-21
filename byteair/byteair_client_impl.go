package byteair

import (
	"errors"
	"fmt"
	"strings"
	"time"

	. "github.com/byteplus-sdk/sdk-go/byteair/protocol"
	"github.com/byteplus-sdk/sdk-go/common"
	commonprotocol "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
)

const DefaultPredictScene = "default"

var (
	errMsgFormat    = "Only can receive max to %d items in one request"
	TooManyItemsErr = errors.New(fmt.Sprintf(errMsgFormat, MaxImportItemCount))
)

type clientImpl struct {
	cCli    common.Client
	hCaller *HttpCaller
	gu      *byteairURL
	hostAva *HostAvailabler
}

func (c *clientImpl) Release() {
	c.hostAva.Shutdown()
}

func (c *clientImpl) GetOperation(request *commonprotocol.GetOperationRequest,
	opts ...option.Option) (*commonprotocol.OperationResponse, error) {
	return c.cCli.GetOperation(request, opts...)
}

func (c *clientImpl) ListOperations(request *commonprotocol.ListOperationsRequest,
	opts ...option.Option) (*commonprotocol.ListOperationsResponse, error) {
	return c.cCli.ListOperations(request, opts...)
}

func (c *clientImpl) WriteData(dataList []map[string]interface{}, topic string,
	opts ...option.Option) (*WriteResponse, error) {
	if len(dataList) > MaxWriteItemCount {
		logs.Warn("[ByteplusSDK][WriteData] item count more than '{}'", MaxWriteItemCount)
		if len(dataList) > MaxImportItemCount {
			return nil, TooManyItemsErr
		}
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

func (c *clientImpl) Done(dateList []time.Time,
	topic string, opts ...option.Option) (*DoneResponse, error) {
	var dates []*Date
	if len(dateList) == 0 {
		previousDay := time.Now().Add(-24 * time.Hour)
		dates = c.appendDoneDate(dates, previousDay)
	} else {
		for _, date := range dateList {
			dates = c.appendDoneDate(dates, date)
		}
	}
	urlFormat := c.gu.doneURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	request := &DoneRequest{
		DataDates: dates,
	}
	response := &DoneResponse{}
	err := c.hCaller.DoPbRequest(url, request, response, option.Conv2Options(opts...))
	if err != nil {
		return nil, err
	}
	logs.Debug("[Done] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) appendDoneDate(dates []*Date,
	date time.Time) []*Date {
	return append(dates, &Date{
		Year:  int32(date.Year()),
		Month: int32(date.Month()),
		Day:   int32(date.Day()),
	})
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
	err := c.hCaller.DoPbRequest(url, request, response, options)
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
