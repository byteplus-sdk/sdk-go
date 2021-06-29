package retail

import (
	"github.com/byteplus-sdk/sdk-go/core/option"
	. "github.com/byteplus-sdk/sdk-go/retail/protocol"
)

type Client interface {
	// WriteUsers
	//
	// Writes at most 100 users at a time. Exceeding 100 in a request protocol.results protocol.in
	// a rejection. One can use this to upload new users, or update existing
	// users (by providing all the fields).
	WriteUsers(request *WriteUsersRequest, opts ...option.Option) (*WriteUsersResponse, error)

	// ImportUsers
	//
	// Bulk import of Users.
	//
	// `Operation.response` is of type ImportUsersResponse. Note that it is
	// possible for a subset of the items to be successfully inserted.
	// Operation.metadata is of type Metadata.
	// This call returns immediately after the server finishes the
	// preliminary validations and persists the request. The caller should
	// keep polling `OperationResponse.operation.name` using `GetOperation`
	// call below to check the status.
	// Note: This can also be used to update the existing data by providing the
	// existing ids. In this case, please make sure you provide all fields.
	ImportUsers(request *ImportUsersRequest, opts ...option.Option) (*OperationResponse, error)

	// WriteProducts
	//
	// Writes at most 100 products at a time. Exceeding 100 in a request protocol.protocol.results
	// in a rejection.
	// One can use this to upload new products, or update existing products (by
	// providing all the fields).  Deleting a product is unsupported. One can
	// update the existing product by
	// setting `product.is_recommendable` to False.
	WriteProducts(request *WriteProductsRequest, opts ...option.Option) (*WriteProductsResponse, error)

	// ImportProducts
	//
	// Bulk import of Products.
	//
	// `Operation.response` is of type ImportUsersResponse. Note that it is
	// possible for a subset of the items to be successfully inserted.
	// Operation.metadata is of type Metadata.
	// This call returns immediately after the server finishes the preliminary
	// validations and persists the request.  The caller should keep polling
	// `OperationResponse.operation.name` using `GetOperation` call below to
	// check the status.
	// Note: This can also be used to update the existing data by providing the
	// existing ids. In this case, please make sure you provide all fields.
	ImportProducts(request *ImportProductsRequest, opts ...option.Option) (*OperationResponse, error)

	// WriteUserEvents
	//
	// Writes at most 100 UserEvents at a time. Exceeding 100 in a request
	// results in a rejection. One should use this to upload new realtime
	// UserEvents.  Note: This is processing realtime data, so we won't dedupe
	// the requests.
	// Please make sure the requests are deduplicated before sending over.
	WriteUserEvents(request *WriteUserEventsRequest, opts ...option.Option) (*WriteUserEventsResponse, error)

	//ImportUserEvents
	//
	// Bulk import of User events.
	//
	// `Operation.response` is of type ImportUsersResponse. Note that it is
	// possible for a subset of the items to be successfully inserted.
	// Operation.metadata is of type Metadata.
	// This call returns immediately after the server finishes the preliminary
	// validations and persists the request.  The caller should keep polling
	// `OperationResponse.operation.name` using `GetOperation` call below to
	// check the status.
	// Please make sure the requests are deduplicated before sending over.
	ImportUserEvents(request *ImportUserEventsRequest, opts ...option.Option) (*OperationResponse, error)

	// GetOperation
	//
	// Gets the operation of a previous long running call.
	GetOperation(request *GetOperationRequest, opts ...option.Option) (*OperationResponse, error)

	// ListOperations
	//
	// Lists operations that match the specified filter in the request.
	ListOperations(request *ListOperationsRequest, opts ...option.Option) (*ListOperationsResponse, error)

	// Predict
	//
	// Gets the list of products (ranked).
	// The updated user data will take effect in 24 hours.
	// The updated product data will take effect in 30 mins.
	// Depending how (realtime or batch) the UserEvents are sent back, it will
	// be fed into the models and take effect after that.
	Predict(request *PredictRequest, scene string, opts ...option.Option) (*PredictResponse, error)

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
	AckServerImpressions(request *AckServerImpressionsRequest,
		opts ...option.Option) (*AckServerImpressionsResponse, error)
}
