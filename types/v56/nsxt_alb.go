package types

// AlbVsHttpRequestRules wraps []AlbVsHttpRequestRule into a body that is suitable for performing
// API requests
type AlbVsHttpRequestRules struct {
	Values []AlbVsHttpRequestRule `json:"values"`
}

// AlbVsHttpRequestRule defines a single ALB Virtual Service HTTP Request rule
type AlbVsHttpRequestRule struct {
	// Name of the rule
	Name string `json:"name"`
	// Whether the rule is active or not
	Active bool `json:"active"`
	// Whether to enable logging with headers on rule match or not
	Logging bool `json:"logging"`
	// MatchCriteria for the HTTP Request
	MatchCriteria AlbVsHttpRequestAndSecurityRuleMatchCriteria `json:"matchCriteria"`
	// HTTP header rewrite action
	// It can be configured in combination with rewrite URL action
	HeaderActions []*AlbVsHttpRequestRuleHeaderActions `json:"headerActions,omitempty"`
	// HTTP redirect action
	// It cannot be configured in combination with other actions
	RedirectAction *AlbVsHttpRequestRuleRedirectAction `json:"redirectAction,omitempty"`
	// HTTP request URL rewrite action
	// It can be configured in combination with multiple header actions
	RewriteURLAction *AlbVsHttpRequestRuleRewriteURLAction `json:"rewriteUrlAction,omitempty"`
}

// AlbVsHttpRequestAndSecurityRuleMatchCriteria defines match criteria for HTTP Request and Security Rules
type AlbVsHttpRequestAndSecurityRuleMatchCriteria struct {
	// Client IP addresses
	ClientIPMatch *AlbVsHttpRequestRuleClientIPMatch `json:"clientIpMatch,omitempty"`
	// Virtual service ports
	ServicePortMatch *AlbVsHttpRequestRuleServicePortMatch `json:"servicePortMatch,omitempty"`
	// HTTP methods such as GET, PUT, DELETE, POST etc.
	MethodMatch *AlbVsHttpRequestRuleMethodMatch `json:"methodMatch,omitempty"`
	// Configure request paths
	Protocol string `json:"protocol,omitempty"`
	// Configure request paths
	PathMatch *AlbVsHttpRequestRulePathMatch `json:"pathMatch,omitempty"`
	// HTTP request query strings in key=value format
	QueryMatch []string `json:"queryMatch,omitempty"`
	// HTTP request headers
	HeaderMatch []AlbVsHttpRequestRuleHeaderMatch `json:"headerMatch,omitempty"`
	// HTTP cookies
	CookieMatch *AlbVsHttpRequestRuleCookieMatch `json:"cookieMatch,omitempty"`
}

// AlbVsHttpRequestRuleClientIPMatch defines match criteria for Client IP addresses
type AlbVsHttpRequestRuleClientIPMatch struct {
	// MatchCriteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.
	MatchCriteria string `json:"matchCriteria"`
	// Either a single IP address, a range of IP addresses or a network CIDR. Must contain at least
	// one item.
	Addresses []string `json:"addresses"`
}

// AlbVsHttpRequestRuleServicePortMatch defines match criteria based on service ports
type AlbVsHttpRequestRuleServicePortMatch struct {
	// MatchCriteria to use for port matching the HTTP request. Options - IS_IN, IS_NOT_IN.
	MatchCriteria string `json:"matchCriteria"`
	// Listening TCP ports. Allowed values are 1-65535.
	Ports []int `json:"ports"`
}

// AlbVsHttpRequestRuleMethodMatch defines match criteria based on HTTP Methods
type AlbVsHttpRequestRuleMethodMatch struct {
	// MatchCriteria to use for matching the method in the HTTP request. Options - IS_IN, IS_NOT_IN.
	MatchCriteria string `json:"matchCriteria"`
	// HTTP methods to match.
	// Options - GET, PUT, POST, DELETE, HEAD, OPTIONS, TRACE, CONNECT, PATCH, PROPFIND, PROPPATCH,
	// MKCOL, COPY, MOVE, LOCK, UNLOCK
	Methods []string `json:"methods"`
}

// AlbVsHttpRequestRulePathMatch defines match criteria based on request path
type AlbVsHttpRequestRulePathMatch struct {
	// MatchCriteria to use for matching the path in the HTTP request URI.
	// Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH,
	// DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH
	MatchCriteria string `json:"matchCriteria"`
	// String values to match the path
	MatchStrings []string `json:"matchStrings"`
}

// AlbVsHttpRequestRuleHeaderMatch defines match criteria based on request headers
type AlbVsHttpRequestRuleHeaderMatch struct {
	// MatchCriteria to use for matching headers and cookies in the HTTP request amd response.
	// Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS,
	// DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL
	MatchCriteria string `json:"matchCriteria"`
	// Name of the HTTP header whose value is to be matched
	Key string `json:"key"`
	// String values to match for an HTTP header
	Value []string `json:"value"`
}

// AlbVsHttpRequestRuleCookieMatch defines match criteria based on cookies and their values
type AlbVsHttpRequestRuleCookieMatch struct {
	// MatchCriteria to use for matching cookies in the HTTP request.
	// Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS,
	// DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL
	MatchCriteria string `json:"matchCriteria"`
	// Name of the HTTP cookie whose value is to be matched
	Key string `json:"key"`
	// String values to match for an HTTP cookie. Must be fewer than 10240 characters.
	Value string `json:"value"`
}

// AlbVsHttpRequestRuleHeaderActions allows to modify request headers
type AlbVsHttpRequestRuleHeaderActions struct {
	// Action for the chosen header
	// Options - ADD, REMOVE, REPLACE
	Action string `json:"action"`
	// Name of HTTP header
	Name string `json:"name"`
	// Value of HTTP header
	Value string `json:"value"`
}

// AlbVsHttpRequestRuleRedirectAction allows to redirect incoming queries
type AlbVsHttpRequestRuleRedirectAction struct {
	// Host to which redirect the request. Default is the original host
	Host string `json:"host"`
	// Keep or drop the query of the incoming request URI in the redirected URI
	KeepQuery bool `json:"keepQuery"`
	// Path to which redirect the request. Default is the original path
	Path string `json:"path"`
	// Port to which redirect the request. Default is 80 for HTTP and 443 for HTTPS protocol
	Port *int `json:"port,omitempty"`
	// HTTP or HTTPS protocol
	Protocol string `json:"protocol"`
	// One of the redirect status codes - 301, 302, 307
	StatusCode int `json:"statusCode"`
}

// AlbVsHttpRequestRuleRewriteURLAction allows to modify URLs for incoming requests
type AlbVsHttpRequestRuleRewriteURLAction struct {
	// Host to use for the rewritten URL. If omitted, the existing host will be used.
	Host string `json:"host"`
	// Path to use for the rewritten URL. If omitted, the existing path will be used.
	Path string `json:"path"`
	// Query string to use or append to the existing query string in the rewritten URL.
	Query string `json:"query,omitempty"`
	// Whether or not to keep the existing query string when rewriting the URL. Defaults to true.
	KeepQuery bool `json:"keepQuery"`
}

// AlbVsHttpResponseRules wraps []AlbVsHttpResponseRule into a body that is suitable for performing
// API requests
type AlbVsHttpResponseRules struct {
	Values []AlbVsHttpResponseRule `json:"values"`
}

// AlbVsHttpResponseRule defines a single ALB Virtual Service HTTP Response rule
type AlbVsHttpResponseRule struct {
	// Name of the rule
	Name string `json:"name"`
	// Whether the rule is active or not
	Active bool `json:"active"`
	// Whether to enable logging with headers on rule match or not
	Logging bool `json:"logging"`
	// MatchCriteria for the HTTP Request
	MatchCriteria AlbVsHttpResponseRuleMatchCriteria `json:"matchCriteria"`
	// HTTP header rewrite action
	// It can be configured in combination with rewrite URL action
	HeaderActions []*AlbVsHttpRequestRuleHeaderActions `json:"headerActions"`
	// HTTP location header rewrite action
	RewriteLocationHeaderAction *AlbVsHttpRespRuleRewriteLocationHeaderAction `json:"rewriteLocationHeaderAction"`
}

// AlbVsHttpResponseRuleMatchCriteria combines all possible match criteria for HTTP Security Rules
type AlbVsHttpResponseRuleMatchCriteria struct {
	// Client IP addresses.
	ClientIPMatch *AlbVsHttpRequestRuleClientIPMatch `json:"clientIpMatch,omitempty"`
	// Virtual service ports.
	ServicePortMatch *AlbVsHttpRequestRuleServicePortMatch `json:"servicePortMatch,omitempty"`
	// HTTP methods such as GET, PUT, DELETE, POST etc.
	MethodMatch *AlbVsHttpRequestRuleMethodMatch `json:"methodMatch,omitempty"`
	// Configure request paths.
	Protocol string `json:"protocol,omitempty"`
	// Configure request paths.
	PathMatch *AlbVsHttpRequestRulePathMatch `json:"pathMatch,omitempty"`
	// HTTP request query strings in key=value format.
	QueryMatch []string `json:"queryMatch,omitempty"`
	// HTTP cookies.
	CookieMatch *AlbVsHttpRequestRuleCookieMatch `json:"cookieMatch,omitempty"`

	// Defines match criteria based on response location header
	LocationHeaderMatch *AlbVsHttpResponseLocationHeaderMatch `json:"locationHeaderMatch,omitempty"`
	// Defines match criteria based on the request headers
	RequestHeaderMatch []AlbVsHttpRequestRuleHeaderMatch `json:"requestHeaderMatch,omitempty"`
	// Defines match criteria based on the response headers
	ResponseHeaderMatch []AlbVsHttpRequestRuleHeaderMatch `json:"responseHeaderMatch,omitempty"`
	// Defines match criteria based on response status codes
	StatusCodeMatch *AlbVsHttpRuleStatusCodeMatch `json:"statusCodeMatch,omitempty"`
}

// AlbVsHttpResponseLocationHeaderMatch defines match criteria for
type AlbVsHttpResponseLocationHeaderMatch struct {
	// MatchCriteria to use for matching the path in the HTTP request URI
	// Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH,
	// DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH
	MatchCriteria string `json:"matchCriteria"`
	// String values to match for an HTTP header
	Value []string `json:"value"`
}

// AlbVsHttpRuleStatusCodeMatch defines criteria for matching response status code
type AlbVsHttpRuleStatusCodeMatch struct {
	// MatchCriteria to use for matching the HTTP response status code
	// Options - IS_IN, IS_NOT_IN
	MatchCriteria string `json:"matchCriteria"`
	// StatusCodes contains a list of single status codes or ranges of status codes
	StatusCodes []string `json:"statusCodes"`
}

// AlbVsHttpRespRuleRewriteLocationHeaderAction allows to rewrite Location header for HTTP Responses
type AlbVsHttpRespRuleRewriteLocationHeaderAction struct {
	// Protocol is HTTP or HTTPS
	Protocol string `json:"protocol"`
	// Host to which redirect the request. Default is the original host
	Host string `json:"host"`
	// Port to which redirect the request. Default is 80 for HTTP and 443 for HTTPS protocol
	Port *int `json:"port,omitempty"`
	// Path to which redirect the request. Default is the original path
	Path string `json:"path"`
	// Keep or drop the query of the incoming request URI in the redirected URI
	KeepQuery bool `json:"keepQuery"`
}

// AlbVsHttpSecurityRules wraps []AlbVsHttpSecurityRule into a body that is suitable for performing
// API requests
type AlbVsHttpSecurityRules struct {
	Values []AlbVsHttpSecurityRule `json:"values"`
}

// AlbVsHttpSecurityRule defines a single ALB Virtual Service HTTP Security rule
type AlbVsHttpSecurityRule struct {
	// Name of the rule. Must be non-blank and fewer than 1000 characters.
	Name string `json:"name"`
	// Whether the rule is active or not.
	Active bool `json:"active"`
	// Whether to enable logging with headers on rule match or not.
	Logging       bool                                         `json:"logging"`
	MatchCriteria AlbVsHttpRequestAndSecurityRuleMatchCriteria `json:"matchCriteria"`

	// Defines the action to apply rate limit on incoming requests. It consists of rate limiting
	// properties and one of the actions to execute upon reaching rate limit. If not actions are
	// provided, rate limiting will only be reported.
	RateLimitAction *AlbVsHttpSecurityRuleRateLimitAction `json:"rateLimitAction,omitempty"`

	// Action to redirect the incoming request to HTTPS. It cannot be configured in combination with other actions.
	RedirectToHTTPSAction *AlbVsHttpSecurityRuleRedirectToHTTPSAction `json:"redirectToHttpsAction,omitempty"`

	// Action to send a local HTTP response. It cannot be configured in combination with other actions.
	LocalResponseAction *AlbVsHttpSecurityRuleRateLimitLocalResponseAction `json:"localResponseAction,omitempty"`

	// AllowOrCloseConnectionAction is an action to allow the incoming request or close the
	// connection. It cannot be configured in combination with other actions. Allowed values are:
	// * ALLOW - Allow the incoming request.
	// * CLOSE - Close the incoming connection.
	AllowOrCloseConnectionAction string `json:"allowOrCloseConnectionAction,omitempty"`
}

// AlbVsHttpSecurityRuleRateLimitAction defines action to apply rate limit on incoming requests
type AlbVsHttpSecurityRuleRateLimitAction struct {
	// Action to close the incoming connection. Only allowed value is CLOSE. It cannot be configured in combination with other actions.
	CloseConnectionAction string `json:"closeConnectionAction,omitempty"`
	// Maximum number of connections, requests or packets permitted each period. Allowed values are 1-1000000000.
	Count int `json:"count,omitempty"`
	// Action to send a local HTTP response. It cannot be configured in combination with other actions.
	LocalResponseAction *AlbVsHttpSecurityRuleRateLimitLocalResponseAction `json:"localResponseAction,omitempty"`
	// Time value in seconds to enforce rate count. Allowed values are 1-1000000000. Unit is Second.
	Period int `json:"period,omitempty"`
	// Action to redirect the HTTP request. It cannot be configured in combination with other actions.
	RedirectAction *AlbVsHttpRequestRuleRedirectAction `json:"redirectAction,omitempty"`
}

// AlbVsHttpSecurityRuleRateLimitLocalResponseAction defines action to send a local HTTP response
type AlbVsHttpSecurityRuleRateLimitLocalResponseAction struct {
	// Content to be used in the local HTTP response body.
	Content string `json:"content"`
	// Mime-type of the response content.
	ContentType string `json:"contentType"`
	// Status code of the response. Options - 200, 204, 403, 404, 429, 501.
	StatusCode int `json:"statusCode"`
}

// AlbVsHttpSecurityRuleRedirectToHTTPSAction to redirect the incoming request to HTTPS
type AlbVsHttpSecurityRuleRedirectToHTTPSAction struct {
	// Prt which the request should be redirected
	Port int `json:"port"`
}
