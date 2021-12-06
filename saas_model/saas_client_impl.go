package saas_model

import (
	"errors"
	"fmt"
	"github.com/byteplus-sdk/sdk-go/common"
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	. "github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/byteplus-sdk/sdk-go/core/option"
	generalpb "github.com/byteplus-sdk/sdk-go/general/protocol"
	. "github.com/byteplus-sdk/sdk-go/retail/protocol"
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

func (c *clientImpl) WriteData(writeRequest *generalpb.WriteDataRequest, topic string,
	opts ...option.Option) (*generalpb.WriteResponse, error) {
	if len(writeRequest.GetDatas()) > MaxWriteItemCount {
		logs.Warn("[ByteplusSDK][WriteData] item count more than '{}'", MaxWriteItemCount)
		if len(writeRequest.GetDatas()) > MaxImportItemCount {
			return nil, writeTooManyErr
		}
	}
	urlFormat := c.su.writeDataURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	response := &generalpb.WriteResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, writeRequest, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteData] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) ImportData(importRequest *generalpb.ImportDataRequest,
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

func (c *clientImpl) Done(dateList []time.Time,
	topic string, opts ...option.Option) (*generalpb.DoneResponse, error) {
	var dateMaps []map[string]string
	if len(dateList) == 0 {
		previousDay := time.Now().Add(-24 * time.Hour)
		dateMaps = c.appendDoneDate(dateMaps, previousDay)
	} else {
		for _, date := range dateList {
			dateMaps = c.appendDoneDate(dateMaps, date)
		}
	}
	urlFormat := c.su.doneURLFormat
	url := strings.ReplaceAll(urlFormat, "{}", topic)
	response := &generalpb.DoneResponse{}
	err := c.hCaller.DoJsonRequest(url, dateMaps, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[Done] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) appendDoneDate(dateMaps []map[string]string,
	date time.Time) []map[string]string {
	dateMap := map[string]string{"partition_date": date.Format("20060102")}
	return append(dateMaps, dateMap)
}

func (c *clientImpl) Predict(request *PredictRequest, modelId string,
	opts ...option.Option) (*PredictResponse, error) {
	url := strings.ReplaceAll(c.su.predictURLFormat, "{}", modelId)
	response := &PredictResponse{}
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

func (c *clientImpl) AckServerImpressions(request *AckServerImpressionsRequest,
	modelId string, opts ...option.Option) (*AckServerImpressionsResponse, error) {
	url := strings.ReplaceAll(c.su.ackImpressionURLFormat, "{}", modelId)
	response := &AckServerImpressionsResponse{}
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
