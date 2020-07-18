package types

import (
	"encoding/json"
	"fmt"
)

// CloudApiPageValues is a slice of json.RawMessage. json.RawMessage itself allows to partially marshal responses and is
// used to decouple API paging handling from particular returned types.
type CloudApiPageValues []json.RawMessage

// CloudApiPages unwraps pagination for "Get All" endpoints in CloudAPI. It uses a type "CloudApiPageValues" for values
// which are kept int []json.RawMessage. json.RawMessage helps to decouple marshalling paging related information from
// exact type related information. Paging can be handled dynamically this way while values can be marshaled into exact
// types.
type CloudApiPages struct {
	// ResultTotal reports total results available
	ResultTotal int `json:"resultTotal,omitempty"`
	// PageCount reports total result pages available
	PageCount int `json:"pageCount,omitempty"`
	// Page reports current page of result
	Page int `json:"page,omitempty"`
	// PageSize reports pagesize
	PageSize int `json:"pageSize,omitempty"`
	// Associations ...
	Associations interface{} `json:"associations,omitempty"`
	// Values holds types depending on the endpoint therefore `json.RawMessage` is used to dynamically unmarshal into
	// specific type as required
	Values json.RawMessage `json:"values,omitempty"`
}

// AccumulatePageResponses helps to accumulate Raw JSON objects during a pagination query
type AccumulatePageResponses []json.RawMessage

// OpenApiError helpes to marshal and provider meaningful `Error` for
type OpenApiError struct {
	MinorErrorCode string `json:"minorErrorCode"`
	Message        string `json:"message"`
	StackTrace     string `json:"stackTrace"`
}

// Error method implements Go's default `error` interface for CloudAPI errors formats them for human readable output.
func (openApiError OpenApiError) Error() string {
	return fmt.Sprintf("%s - %s", openApiError.MinorErrorCode, openApiError.Message)
}

// ErrorWithStack is the same as `Error()`, but also includes stack trace returned by API which is usually lengthy.
func (openApiError OpenApiError) ErrorWithStack() string {
	return fmt.Sprintf("%s - %s. Stack: %s", openApiError.MinorErrorCode, openApiError.Message,
		openApiError.StackTrace)
}
