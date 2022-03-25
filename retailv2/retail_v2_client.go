package retailv2

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core/option"
	. "github.com/byteplus-sdk/sdk-go/retailv2/protocol"
)

type Client interface {
	common.Client

	// Release
	// release resources
	Release()

	// WriteUsers
	//
	// Writes at most 2000 users at a time. Exceeding 2000 in a request protocol.results protocol.in
	// a rejection. One can use this to upload new users, or update existing
	// users (by providing all the fields).
	WriteUsers(request *WriteUsersRequest, opts ...option.Option) (*WriteUsersResponse, error)

	// WriteProducts
	//
	// Writes at most 2000 products at a time. Exceeding 2000 in a request protocol.protocol.results
	// in a rejection.
	// One can use this to upload new products, or update existing products (by
	// providing all the fields).  Deleting a product is unsupported. One can
	// update the existing product by
	// setting `product.is_recommendable` to False.
	WriteProducts(request *WriteProductsRequest, opts ...option.Option) (*WriteProductsResponse, error)

	// WriteUserEvents
	//
	// Writes at most 2000 UserEvents at a time. Exceeding 2000 in a request
	// results in a rejection. One should use this to upload new realtime
	// UserEvents.  Note: This is processing realtime data, so we won't dedupe
	// the requests.
	// Please make sure the requests are deduplicated before sending over.
	WriteUserEvents(request *WriteUserEventsRequest, opts ...option.Option) (*WriteUserEventsResponse, error)

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
