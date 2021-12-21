package byteair

import (
	"time"

	. "github.com/byteplus-sdk/sdk-go/byteair/protocol"
	"github.com/byteplus-sdk/sdk-go/common"
	commonprotocol "github.com/byteplus-sdk/sdk-go/common/protocol"
	"github.com/byteplus-sdk/sdk-go/core/option"
)

type Client interface {
	common.Client

	// Release
	// release resources
	Release()

	// WriteData
	//
	// Writes at most 100 data at a time. Exceeding 100 in a request results in
	// a rejection. One can use this to upload new data, or update existing
	// data (by providing all the fields, some data type not support update, e.g. user event).
	WriteData(dataList []map[string]interface{}, topic string,
		opts ...option.Option) (*WriteResponse, error)

	// Done
	//
	// When the data of a day is imported completely,
	// you should notify bytedance through `done` method,
	// then bytedance will start handling the data in this day
	// @param dateList, optional, if dataList is empty, indicate target date is previous day
	Done(dateList []time.Time, topic string, opts ...option.Option) (*DoneResponse, error)

	// GetOperation
	//
	// Gets the operation of a previous long running call.
	GetOperation(request *commonprotocol.GetOperationRequest, opts ...option.Option) (*commonprotocol.OperationResponse, error)

	// ListOperations
	//
	// Lists operations that match the specified filter in the request.
	ListOperations(request *commonprotocol.ListOperationsRequest, opts ...option.Option) (*commonprotocol.ListOperationsResponse, error)

	// Predict
	//
	// Gets the list of products (ranked).
	// The updated user data will take effect in 24 hours.
	// The updated product data will take effect in 30 mins.
	// Depending how (realtime or batch) the UserEvents are sent back, it will
	// be fed into the models and take effect after that.
	Predict(request *PredictRequest, opts ...option.Option) (*PredictResponse, error)

	// Callback
	//
	// Sends back the actual product list shown to the users based on the
	// customized changes from `PredictResponse`.
	// example: our Predict call returns the list of items [1, 2, 3, 4].
	// Your custom logic have decided that product 3 has been sold out and
	// product 10 needs to be inserted before 2 based on some promotion rules,
	// the AckServerImpressionsRequest content items should looks like
	// [
	//   {id:1, extra: "{\"reason\": \"kept\"}", pos:1},
	//   {id:10, extra: "{\"reason\": \"inserted\"}", pos:2},
	//   {id:2, extra: "{\"reason\": \"kept\"}", pos:3},
	//   {id:4, extra: "{\"reason\": \"kept\"}", pos:4},
	//   {id:3, extra: "{\"reason\": \"filtered\"}", pos:0},
	// ].
	Callback(request *CallbackRequest, opts ...option.Option) (*CallbackResponse, error)
}
