package saas

import (
	"errors"
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	"github.com/byteplus-sdk/sdk-go/saas/protocol"
	"strings"
	"time"
)

var (
	writeMsgFormat  = "Only can receive max to %d items in one write request"
	writeTooManyErr = errors.New(fmt.Sprintf(writeMsgFormat, MaxWriteItemCount))
)

type clientImpl struct {
	cCli    common.Client
	hCaller *HttpCaller
	su      *saasURL
	hostAva *HostAvailabler
}

func (c *clientImpl) Release() {
	c.hostAva.Shutdown()
}

func (c *clientImpl) WriteData(writeRequest *protocol.WriteDataRequest, topic string,
	opts ...option.Option) (*protocol.WriteResponse, error) {
	if len(writeRequest.GetDatas()) > MaxWriteItemCount {
		logs.Warn("[ByteplusSDK][WriteData] item count more than '{}'", MaxWriteItemCount)
		if len(writeRequest.GetDatas()) > MaxImportItemCount {
			return nil, writeTooManyErr
		}
	}
	urlFormat := c.su.writeDataURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	response := &protocol.WriteResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, writeRequest, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteData] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ImportData(importRequest *protocol.ImportDataRequest,
	topic string, opts ...option.Option) (*OperationResponse, error) {
	urlFormat := c.su.importDataURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	response := &OperationResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, importRequest, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[ImportData] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) Done(request *protocol.DoneRequest,
	topic string, opts ...option.Option) (*protocol.DoneResponse, error) {
	if len(request.DataDates) == 0 {
		previousDay := time.Now().Add(-24 * time.Hour)
		date := convTime2Date(previousDay)
		request.DataDates = append(request.DataDates, &date)
	}
	urlFormat := c.su.doneURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	response := &protocol.DoneResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[Done] rsp:\n%s\n", response)
	return response, nil
}

func convTime2Date(day time.Time) protocol.Date {
	return protocol.Date{
		Year:  int32(day.Year()),
		Month: int32(day.Month()),
		Day:   int32(day.Day()),
	}
}

func (c *clientImpl) Predict(request *protocol.PredictRequest, modelId string,
	opts ...option.Option) (*protocol.PredictResponse, error) {
	url := strings.ReplaceAll(c.su.predictURLFormat, "{}", modelId)
	response := &protocol.PredictResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[Predict] rsp:\n%s\n", response)
	return response, nil
}

func addSaasFlag(opts []option.Option) []option.Option {
	if len(opts) == 0 {
		opts = []option.Option{withSaasHeader()}
		return opts
	}
	return append(opts, withSaasHeader())
}

func withSaasHeader() option.Option {
	const (
		HTTPHeaderServerFrom = "Server-From"
		SaasFlag             = "saas"
	)
	return func(opt *option.Options) {
		if len(opt.Headers) == 0 {
			opt.Headers = map[string]string{HTTPHeaderServerFrom: SaasFlag}
			return
		}
		opt.Headers[HTTPHeaderServerFrom] = SaasFlag
	}
}

func (c *clientImpl) AckServerImpressions(request *protocol.AckServerImpressionsRequest,
	modelId string, opts ...option.Option) (*protocol.AckServerImpressionsResponse, error) {
	url := strings.ReplaceAll(c.su.ackImpressionURLFormat, "{}", modelId)
	response := &protocol.AckServerImpressionsResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[AckImpressions] rsp:\n%s\n", response)
	return response, nil
}

// GetOperation
//
// Gets the operation of a previous long running call.
func (c *clientImpl) GetOperation(request *GetOperationRequest,
	opts ...option.Option) (*OperationResponse, error) {
	return c.cCli.GetOperation(request, opts...)
}

// ListOperations
//
// Lists operations that match the specified filter in the request.
func (c *clientImpl) ListOperations(request *ListOperationsRequest,
	opts ...option.Option) (*ListOperationsResponse, error) {
	return c.cCli.ListOperations(request, opts...)
}
