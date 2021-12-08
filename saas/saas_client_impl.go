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
)

var (
	writeMsgFormat  = "Only can receive max to %d items in one write request"
	writeTooManyErr = errors.New(fmt.Sprintf(writeMsgFormat, MaxWriteItemCount))
)

const (
	dataInitOptionCount    = 2
	predictInitOptionCount = 1
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

func checkProjectIdAndModelId(projectId string, modelId string) error {
	const (
		errMsgFormat      = "%s,field can not empty"
		errFieldProjectId = "projectId"
		errFieldModelId   = "modelId"
	)
	if projectId != "" && modelId != "" {
		return nil
	}
	emptyParams := make([]string, 0)
	if projectId == "" {
		emptyParams = append(emptyParams, errFieldProjectId)
	}
	if modelId == "" {
		emptyParams = append(emptyParams, errFieldModelId)
	}
	return errors.New(fmt.Sprintf(errMsgFormat, strings.Join(emptyParams, ",")))
}

func checkProjectIdAndStage(projectId string, stage string) error {
	const (
		errMsgFormat      = "%s,field can not empty"
		errFieldProjectId = "projectId"
		errFieldStage     = "stage"
	)
	if projectId != "" && stage != "" {
		return nil
	}
	emptyParams := make([]string, 0)
	if projectId == "" {
		emptyParams = append(emptyParams, errFieldProjectId)
	}
	if stage == "" {
		emptyParams = append(emptyParams, errFieldStage)
	}
	return errors.New(fmt.Sprintf(errMsgFormat, strings.Join(emptyParams, ",")))
}

func addSaasFlag(opts []option.Option) []option.Option {
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

func (c *clientImpl) doWrite(request *protocol.WriteDataRequest, url string, opts ...option.Option) (*protocol.WriteResponse, error) {
	if err := checkProjectIdAndStage(request.ProjectId, request.Stage); err != nil {
		return nil, err
	}
	if len(request.GetDatas()) > MaxWriteItemCount {
		logs.Warn("[ByteplusSDK][WriteData] item count more than '{}'", MaxWriteItemCount)
		if len(request.GetDatas()) > MaxImportItemCount {
			return nil, writeTooManyErr
		}
	}
	if len(opts) == 0 {
		opts = make([]option.Option, 0, dataInitOptionCount)
	}
	response := &protocol.WriteResponse{}
	opts = append(opts, option.WithStage(request.Stage))
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(url, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[WriteData] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) WriteUsersData(writeRequest *protocol.WriteDataRequest, opts ...option.Option) (*protocol.WriteResponse, error) {
	return c.doWrite(writeRequest, c.su.writeUsersDataURL, opts...)
}

func (c *clientImpl) WriteProductsData(writeRequest *protocol.WriteDataRequest, opts ...option.Option) (*protocol.WriteResponse, error) {
	return c.doWrite(writeRequest, c.su.writeProductsDataURL, opts...)
}

func (c *clientImpl) WriteUserEventsData(writeRequest *protocol.WriteDataRequest, opts ...option.Option) (*protocol.WriteResponse, error) {
	return c.doWrite(writeRequest, c.su.writeUserEventsDataURL, opts...)
}

func (c *clientImpl) Predict(request *protocol.PredictRequest, opts ...option.Option) (*protocol.PredictResponse, error) {
	if err := checkProjectIdAndModelId(request.ProjectId, request.ModelId); err != nil {
		return nil, err
	}
	if len(opts) == 0 {
		opts = make([]option.Option, 0, predictInitOptionCount)
	}
	response := &protocol.PredictResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(c.su.predictURL, request, response, opts...)
	if err != nil {
		return nil, err
	}
	logs.Debug("[Predict] rsp:\n%s\n", response)
	return response, nil
}

func (c *clientImpl) AckServerImpressions(request *protocol.AckServerImpressionsRequest,
	opts ...option.Option) (*protocol.AckServerImpressionsResponse, error) {
	if err := checkProjectIdAndModelId(request.ProjectId, request.ModelId); err != nil {
		return nil, err
	}
	if len(opts) == 0 {
		opts = make([]option.Option, 0, predictInitOptionCount)
	}
	response := &protocol.AckServerImpressionsResponse{}
	opts = addSaasFlag(opts)
	err := c.hCaller.DoPbRequest(c.su.ackImpressionURL, request, response, opts...)
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
