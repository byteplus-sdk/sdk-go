package common

import (
	. "github.com/byteplus-sdk/sdk-go/common/protocol"
	"github.com/byteplus-sdk/sdk-go/core/option"
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
}