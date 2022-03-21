package media

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core/option"
	"github.com/byteplus-sdk/sdk-go/media/protocol"
)

type Client interface {
	common.Client

	// WriteUsers
	//
	// Writes at most 100 users at a time. Exceeding 100 in a request protocol.results protocol.in
	// a rejection. One can use this to upload new users, or update existing
	// users (by providing all the fields).
	WriteUsers(request *protocol.WriteUsersRequest,
		opts ...option.Option) (*protocol.WriteUsersResponse, error)

	// WriteContents
	//
	// Writes at most 2000 contents at a time. Exceeding 2000 in a request protocol.protocol.results
	// in a rejection.
	// One can use this to upload new contents, or update existing contents (by
	// providing all the fields).  Deleting a content is unsupported. One can
	// update the existing content by
	// setting `content.is_recommendable` to False.
	WriteContents(request *protocol.WriteContentsRequest,
		opts ...option.Option) (*protocol.WriteContentsResponse, error)

	// WriteUserEvents
	//
	// Writes at most 2000 UserEvents at a time. Exceeding 2000 in a request
	// results in a rejection. One should use this to upload new realtime
	// UserEvents.  Note: This is processing realtime data, so we won't dedupe
	// the requests.
	// Please make sure the requests are deduplicated before sending over.
	WriteUserEvents(request *protocol.WriteUserEventsRequest,
		opts ...option.Option) (*protocol.WriteUserEventsResponse, error)

	// Release resources
	Release()
}
