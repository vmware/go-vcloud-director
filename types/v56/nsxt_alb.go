package types

type AlbVsHttpRequestRules struct {
	Values []AlbVsHttpRequestRule `json:"values"`
}

type AlbVsHttpRequestRule struct {
	// Name of the rule. Must be non-blank and fewer than 1000 characters.
	Name string `json:"name"`
	// Whether the rule is active or not.
	Active bool `json:"active"`
	// Whether to enable logging with headers on rule match or not.
	Logging       bool                              `json:"logging"`
	MatchCriteria AlbVsHttpRequestRuleMatchCriteria `json:"matchCriteria"`
	// HTTP header rewrite action. It can be configured in combination with rewrite URL action.
	HeaderActions []*AlbVsHttpRequestRuleHeaderActions `json:"headerActions,omitempty"`
	// HTTP redirect action. It cannot be configured in combination with other actions.
	RedirectAction *AlbVsHttpRequestRuleRedirectAction `json:"redirectAction,omitempty"`
	// HTTP request URL rewrite action. It can be configured in combination with multiple header actions.
	RewriteURLAction *AlbVsHttpRequestRuleRewriteURLAction `json:"rewriteUrlAction,omitempty"`
}

type AlbVsHttpRequestRuleMatchCriteria struct {
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
	// HTTP request headers.
	HeaderMatch []AlbVsHttpRequestRuleHeaderMatch `json:"headerMatch,omitempty"`
	// HTTP cookies.
	CookieMatch *AlbVsHttpRequestRuleCookieMatch `json:"cookieMatch,omitempty"`
}

// Client IP addresses.
type AlbVsHttpRequestRuleClientIPMatch struct {
	// Criterion to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.
	MatchCriteria string `json:"matchCriteria"`
	// Either a single IP address, a range of IP addresses or a network CIDR. Must contain at least one item.
	Addresses []string `json:"addresses"`
}

type AlbVsHttpRequestRuleServicePortMatch struct {
	// Criterion to use for port matching the HTTP request. Options - IS_IN, IS_NOT_IN.
	MatchCriteria string `json:"matchCriteria"`
	// Listening TCP ports. Allowed values are 1-65535.
	Ports []int `json:"ports"`
}

type AlbVsHttpRequestRuleMethodMatch struct {
	// Criterion to use for matching the method in the HTTP request. Options - IS_IN, IS_NOT_IN.
	MatchCriteria string `json:"matchCriteria"`
	// HTTP methods to match. Options - GET, PUT, POST, DELETE, HEAD, OPTIONS, TRACE, CONNECT,
	// PATCH, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK.
	Methods []string `json:"methods"`
}

// Configure request paths.
type AlbVsHttpRequestRulePathMatch struct {
	// Criterion to use for matching the path in the HTTP request URI. Options - BEGINS_WITH,
	// DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS,
	// DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH.
	MatchCriteria string `json:"matchCriteria"`
	// String values to match the path.
	MatchStrings []string `json:"matchStrings"`
}

// HTTP request headers.
type AlbVsHttpRequestRuleHeaderMatch struct {
	// Criterion to use for matching headers and cookies in the HTTP request amd response. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL.
	MatchCriteria string `json:"matchCriteria"`
	// String values to match for an HTTP header.
	Value []string `json:"value"`
	// Name of the HTTP header whose value is to be matched. Must be non-blank and fewer than 10240 characters.
	Key string `json:"key"`
}

type AlbVsHttpRequestRuleCookieMatch struct {
	// Criterion to use for matching cookies in the HTTP request. Options - EXISTS, DOES_NOT_EXIST,
	// BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH,
	// EQUALS, DOES_NOT_EQUAL.
	MatchCriteria string `json:"matchCriteria"`
	// Name of the HTTP cookie whose value is to be matched. Must be non-blank and fewer than 10240
	// characters.
	Key string `json:"key"`
	// String values to match for an HTTP cookie. Must be fewer than 10240 characters.
	Value string `json:"value"`
}

type AlbVsHttpRequestRuleHeaderActions struct {
	// One of the following HTTP header actions. Options - ADD, REMOVE, REPLACE.
	Action string `json:"action"`
	// HTTP header name. Must be non-blank and fewer than 10240 characters.
	Name string `json:"name"`
	// HTTP header value. Must be non-blank and fewer than 10240 characters.
	Value string `json:"value"`
}

type AlbVsHttpRequestRuleRedirectAction struct {
	// Host to which redirect the request. Default is the original host.
	Host string `json:"host"`
	// Keep or drop the query of the incoming request URI in the redirected URI.
	KeepQuery bool `json:"keepQuery"`
	// Path to which redirect the request. Default is the original path.
	Path string `json:"path"`
	// Port to which redirect the request. Default is 80 for HTTP and 443 for HTTPS protocol.
	Port int `json:"port"`
	// HTTP or HTTPS protocol. Must be non-blank.
	Protocol string `json:"protocol"`
	// One of the redirect status codes - 301, 302, 307.
	StatusCode int `json:"statusCode"`
}

type AlbVsHttpRequestRuleRewriteURLAction struct {
	// Host to use for the rewritten URL. If omitted, the existing host will be used.
	Host string `json:"host"`
	// Path to use for the rewritten URL. If omitted, the existing path will be used.
	Path string `json:"path"`
	// Query string to use or append to the existing query string in the rewritten URL.
	Query string `json:"query"`
	// Whether or not to keep the existing query string when rewriting the URL. Defaults to true.
	KeepQuery bool `json:"keepQuery"`
}

//////

type AlbVsHttpResponseRules struct {
	Values []AlbVsHttpResponseRule `json:"values"`
}

type AlbVsHttpResponseRule struct {
	Name                        string                                        `json:"name"`
	Active                      bool                                          `json:"active"`
	Logging                     bool                                          `json:"logging"`
	MatchCriteria               AlbVsHttpResponseRuleMatchCriteria            `json:"matchCriteria"`
	HeaderActions               []*AlbVsHttpRequestRuleHeaderActions          `json:"headerActions"`
	RewriteLocationHeaderAction *AlbVsHttpRespRuleRewriteLocationHeaderAction `json:"rewriteLocationHeaderAction"`
}

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

	LocationHeaderMatch *AlbVsHttpResponseLocationHeaderMatch `json:"locationHeaderMatch,omitempty"`
	RequestHeaderMatch  []AlbVsHttpRequestRuleHeaderMatch     `json:"requestHeaderMatch,omitempty"`
	ResponseHeaderMatch []AlbVsHttpRequestRuleHeaderMatch     `json:"responseHeaderMatch,omitempty"`
	StatusCodeMatch     *AlbVsHttpRuleStatusCodeMatch         `json:"statusCodeMatch,omitempty"`
}

type AlbVsHttpResponseLocationHeaderMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Value         []string `json:"value"`
}

type AlbVsHttpRuleStatusCodeMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	StatusCodes   []string `json:"statusCodes"`
}

type AlbVsHttpRespRuleRewriteLocationHeaderAction struct {
	Protocol  string `json:"protocol"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Path      string `json:"path"`
	KeepQuery bool   `json:"keepQuery"`
}

type AlbVsHttpSecurityRules struct {
	Values []AlbVsHttpSecurityRule `json:"values"`
}

type AlbVsHttpSecurityRule struct {
	// Name of the rule. Must be non-blank and fewer than 1000 characters.
	Name string `json:"name"`
	// Whether the rule is active or not.
	Active bool `json:"active"`
	// Whether to enable logging with headers on rule match or not.
	Logging       bool                              `json:"logging"`
	MatchCriteria AlbVsHttpRequestRuleMatchCriteria `json:"matchCriteria"`

	//Defines the action to apply rate limit on incoming requests. It consists of rate limiting
	//properties and one of the actions to execute upon reaching rate limit. If not actions are
	//provided, rate limiting will only be reported.
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
type AlbVsHttpSecurityRuleRateLimitLocalResponseAction struct {
	// Content to be used in the local HTTP response body.
	Content string `json:"content"`
	// Mime-type of the response content.
	ContentType string `json:"contentType"`
	// Status code of the response. Options - 200, 204, 403, 404, 429, 501.
	StatusCode int `json:"statusCode"`
}

type AlbVsHttpSecurityRuleRedirectToHTTPSAction struct {
	Port int `json:"port"`
}
