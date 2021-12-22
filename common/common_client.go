package common

import (
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	"github.com/byteplus-sdk/sdk-go/core/option"
	"time"
)

type Client interface {
	// GetOperation
	//
	// Gets the operation of a previous long running call.
	GetOperation(request *GetOperationRequest, opts ...option.Option) (*OperationResponse, error)

	// ListOperations
	//
	// Lists operations that match the specified filter in the request.
	ListOperations(request *ListOperationsRequest, opts ...option.Option) (*ListOperationsResponse, error)

	// Done
	//
	// When the data of a day is imported completely,
	// you should notify bytedance through `done` method,
	// then bytedance will start handling the data in this day
	// @param dateList, optional, if dataList is empty, indicate target date is previous day
	Done(dateList []time.Time, topic string, opts ...option.Option) (*DoneResponse, error)
}
