package httpwrapper

import (
	"net/http"
	"net/url"
)

type ProblemDetails struct {
	// string providing an URI formatted according to IETF RFC 3986.
	Type string `json:"type,omitempty" yaml:"type" bson:"type" mapstructure:"Type"`
	// A short, human-readable summary of the problem type. It should not change from occurrence to occurrence of the problem.
	Title string `json:"title,omitempty" yaml:"title" bson:"title" mapstructure:"Title"`
	// The HTTP status code for this occurrence of the problem.
	Status int32 `json:"status,omitempty" yaml:"status" bson:"status" mapstructure:"Status"`
	// A human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail,omitempty" yaml:"detail" bson:"detail" mapstructure:"Detail"`
	// string providing an URI formatted according to IETF RFC 3986.
	Instance string `json:"instance,omitempty" yaml:"instance" bson:"instance" mapstructure:"Instance"`
	// A machine-readable application error cause specific to this occurrence of the problem. This IE should be present and provide application-related error information, if available.
	Cause string `json:"cause,omitempty" yaml:"cause" bson:"cause" mapstructure:"Cause"`
	// Description of invalid parameters, for a request rejected due to invalid parameters.
	InvalidParams []InvalidParam `json:"invalidParams,omitempty" yaml:"invalidParams" bson:"invalidParams" mapstructure:"InvalidParams"`
}

type InvalidParam struct {
	// Attribute's name encoded as a JSON Pointer, or header's name.
	Param string `json:"param" yaml:"param" bson:"param" mapstructure:"Param"`
	// A human-readable reason, e.g. \"must be a positive integer\".
	Reason string `json:"reason,omitempty" yaml:"reason" bson:"reason" mapstructure:"Reason"`
}

type Request struct {
	Params map[string]string
	Header http.Header
	Query  url.Values
	Body   interface{}
	URL    *url.URL
}

func NewRequest(req *http.Request, body interface{}) *Request {
	ret := &Request{}
	ret.Query = req.URL.Query()
	ret.Header = req.Header
	ret.Body = body
	ret.Params = make(map[string]string)
	ret.URL = req.URL
	return ret
}

type Response struct {
	Header http.Header
	Status int
	Body   interface{}
}

func NewResponse(code int, h http.Header, body interface{}) *Response {
	ret := &Response{}
	ret.Status = code
	ret.Header = h
	ret.Body = body
	return ret
}
