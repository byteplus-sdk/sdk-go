package saas_model

import (
	"github.com/byteplus-sdk/sdk-go/common"
	commonpb "github.com/byteplus-sdk/sdk-go/common/protocol"
	"github.com/byteplus-sdk/sdk-go/core/option"
	general "github.com/byteplus-sdk/sdk-go/general/protocol"
	. "github.com/byteplus-sdk/sdk-go/retail/protocol"
	"time"
)

type Client interface {
	common.Client

	// Release
	// release resources
	Release()

	// WriteData
	//
	// Writes at most 100 data at a time. Exceeding 100 in a request results in
	// a rejection.Each element of dataList array is a json serialized string of data.
	// One can use this to upload new data, or update existing data (by providing all the fields,
	// some data type not support update, e.g. user event).
	WriteData(writeRequest *general.WriteDataRequest, topic string,
		opts ...option.Option) (*general.WriteResponse, error)

	// ImportData
	//
	// Bulk import of data.
	//
	// `Operation.response` is of type ImportResponse. Note that it is
	// possible for a subset of the items to be successfully inserted.
	// Operation.metadata is of type Metadata.
	// This call returns immediately after the server finishes the
	// preliminary validations and persists the request. The caller should
	// keep polling `OperationResponse.operation.name` using `GetOperation`
	// call below to check the status.
	// Note: This can also be used to update the existing data(some data type not support).
	// In this case, please make sure you provide all fields.
	ImportData(importRequest *general.ImportDataRequest, topic string,
		opts ...option.Option) (*commonpb.OperationResponse, error)

	// Done
	//
	// When the data of a day is imported completely,
	// you should notify bytedance through `done` method,
	// then bytedance will start handling the data in this day
	// @param dateList, optional, if dataList is empty, indicate target date is previous day
	Done(dateList []time.Time, topic string, opts ...option.Option) (*general.DoneResponse, error)

	// Predict
	//
	// Gets the list of products (ranked).
	// The updated user data will take effect in 24 hours.
	// The updated product data will take effect in 30 mins.
	// Depending how (realtime or batch) the UserEvents are sent back, it will
	// be fed into the models and take effect after that.
	Predict(request *PredictRequest, modelId string, opts ...option.Option) (*PredictResponse, error)

	// AckServerImpressions
	//
	// Sends back the actual product list shown to the users based on the
	// customized changes from `PredictResponse`.
	// example: our Predict call returns the list of items [1, 2, 3, 4].
	// Your custom logic have decided that product 3 has been sold out and
	// product 10 needs to be inserted before 2 based on some promotion rules,
	// the AckServerImpressionsRequest content items should looks like
	// [
	//   {id:1, altered_reason: "kept", rank:1},
	//   {id:10, altered_reason: "inserted", rank:2},
	//   {id:2, altered_reason: "kept", rank:3},
	//   {id:4, altered_reason: "kept", rank:4},
	//   {id:3, altered_reason: "filtered", rank:0},
	// ].
	AckServerImpressions(request *AckServerImpressionsRequest, modelId string,
		opts ...option.Option) (*AckServerImpressionsResponse, error)
}
