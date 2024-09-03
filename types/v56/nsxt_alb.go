package types

type EdgeVirtualServiceHttpRequestRule struct {
	Name             string                                            `json:"name"`
	Active           bool                                              `json:"active"`
	Logging          bool                                              `json:"logging"`
	MatchCriteria    EdgeVirtualServiceHttpRequestRuleMatchCriteria    `json:"matchCriteria"`
	HeaderActions    []EdgeVirtualServiceHttpRequestRuleHeaderActions  `json:"headerActions"`
	RedirectAction   EdgeVirtualServiceHttpRequestRuleRedirectAction   `json:"redirectAction"`
	RewriteURLAction EdgeVirtualServiceHttpRequestRuleRewriteURLAction `json:"rewriteUrlAction"`
}

type EdgeVirtualServiceHttpRequestRuleClientIPMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Addresses     []string `json:"addresses"`
}

type EdgeVirtualServiceHttpRequestRuleServicePortMatch struct {
	MatchCriteria string `json:"matchCriteria"`
	Ports         []int  `json:"ports"`
}

type EdgeVirtualServiceHttpRequestRuleMethodMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Methods       []string `json:"methods"`
}

type EdgeVirtualServiceHttpRequestRulePathMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	MatchStrings  []string `json:"matchStrings"`
}

type EdgeVirtualServiceHttpRequestRuleHeaderMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Value         []string `json:"value"`
	Key           string   `json:"key"`
}

type EdgeVirtualServiceHttpRequestRuleCookieMatch struct {
	MatchCriteria string `json:"matchCriteria"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}

type EdgeVirtualServiceHttpRequestRuleMatchCriteria struct {
	ClientIPMatch    EdgeVirtualServiceHttpRequestRuleClientIPMatch    `json:"clientIpMatch"`
	ServicePortMatch EdgeVirtualServiceHttpRequestRuleServicePortMatch `json:"servicePortMatch"`
	MethodMatch      EdgeVirtualServiceHttpRequestRuleMethodMatch      `json:"methodMatch"`
	Protocol         string                                            `json:"protocol"`
	PathMatch        EdgeVirtualServiceHttpRequestRulePathMatch        `json:"pathMatch"`
	QueryMatch       []string                                          `json:"queryMatch"`
	HeaderMatch      []EdgeVirtualServiceHttpRequestRuleHeaderMatch    `json:"headerMatch"`
	CookieMatch      EdgeVirtualServiceHttpRequestRuleCookieMatch      `json:"cookieMatch"`
}

type EdgeVirtualServiceHttpRequestRuleHeaderActions struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

type EdgeVirtualServiceHttpRequestRuleRedirectAction struct {
	StatusCode int `json:"statusCode"`
}

type EdgeVirtualServiceHttpRequestRuleRewriteURLAction struct {
	Host      string `json:"host"`
	Path      string `json:"path"`
	Query     string `json:"query"`
	KeepQuery bool   `json:"keepQuery"`
}

//////

type EdgeVirtualServiceHttpResponseRule struct {
	Name                        string                      `json:"name"`
	Active                      bool                        `json:"active"`
	Logging                     bool                        `json:"logging"`
	MatchCriteria               MatchCriteria               `json:"matchCriteria"`
	HeaderActions               []HeaderActions             `json:"headerActions"`
	RewriteLocationHeaderAction RewriteLocationHeaderAction `json:"rewriteLocationHeaderAction"`
}
type ClientIPMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Addresses     []string `json:"addresses"`
}
type ServicePortMatch struct {
	MatchCriteria string `json:"matchCriteria"`
	Ports         []int  `json:"ports"`
}
type MethodMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Methods       []string `json:"methods"`
}
type PathMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	MatchStrings  []string `json:"matchStrings"`
}
type CookieMatch struct {
	MatchCriteria string `json:"matchCriteria"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}
type LocationHeaderMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Value         []string `json:"value"`
}
type RequestHeaderMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Value         []string `json:"value"`
	Key           string   `json:"key"`
}
type ResponseHeaderMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	Value         []string `json:"value"`
	Key           string   `json:"key"`
}
type StatusCodeMatch struct {
	MatchCriteria string   `json:"matchCriteria"`
	StatusCodes   []string `json:"statusCodes"`
}
type MatchCriteria struct {
	ClientIPMatch       ClientIPMatch         `json:"clientIpMatch"`
	ServicePortMatch    ServicePortMatch      `json:"servicePortMatch"`
	MethodMatch         MethodMatch           `json:"methodMatch"`
	Protocol            string                `json:"protocol"`
	PathMatch           PathMatch             `json:"pathMatch"`
	QueryMatch          []string              `json:"queryMatch"`
	CookieMatch         CookieMatch           `json:"cookieMatch"`
	LocationHeaderMatch LocationHeaderMatch   `json:"locationHeaderMatch"`
	RequestHeaderMatch  []RequestHeaderMatch  `json:"requestHeaderMatch"`
	ResponseHeaderMatch []ResponseHeaderMatch `json:"responseHeaderMatch"`
	StatusCodeMatch     StatusCodeMatch       `json:"statusCodeMatch"`
}
type HeaderActions struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}
type RewriteLocationHeaderAction struct {
	Protocol  string `json:"protocol"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Path      string `json:"path"`
	KeepQuery bool   `json:"keepQuery"`
}
