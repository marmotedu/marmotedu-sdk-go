// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rest

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/marmotedu/component-base/pkg/auth"
	"github.com/marmotedu/component-base/pkg/runtime"

	"github.com/marmotedu/marmotedu-sdk-go/third_party/forked/gorequest"
)

// Request allows for building up a request to a server in a chained fashion.
// Any errors are stored until the end of your call, so you only have to
// check once.
type Request struct {
	c *RESTClient

	timeout time.Duration

	// generic components accessible via method setters
	verb       string
	pathPrefix string
	subpath    string
	params     url.Values
	headers    http.Header

	// structural elements of the request that are part of the IAM API conventions
	// namespace    string
	// namespaceSet bool
	resource     string
	resourceName string
	subresource  string

	// output
	err  error
	body interface{}
}

// NewRequest creates a new request helper object for accessing runtime.Objects on a server.
func NewRequest(c *RESTClient) *Request {
	var pathPrefix string
	if c.base != nil {
		pathPrefix = path.Join("/", c.base.Path, c.versionedAPIPath)
	} else {
		pathPrefix = path.Join("/", c.versionedAPIPath)
	}

	r := &Request{
		c:          c,
		pathPrefix: pathPrefix,
	}

	authMethod := 0

	for _, fn := range []func() bool{c.content.HasBasicAuth, c.content.HasTokenAuth, c.content.HasKeyAuth} {
		if fn() {
			authMethod++
		}
	}

	if authMethod > 1 {
		r.err = fmt.Errorf(
			"username/password or bearer token or secretID/secretKey may be set, but should use only one of them",
		)

		return r
	}

	switch {
	case c.content.HasTokenAuth():
		r.SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.content.BearerToken))
	case c.content.HasKeyAuth():
		tokenString := auth.Sign(c.content.SecretID, c.content.SecretKey, "marmotedu-sdk-go", c.group+".marmotedu.com")
		r.SetHeader("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	case c.content.HasBasicAuth():
		// TODO: get token and set header
		r.SetHeader("Authorization", "Basic "+basicAuth(c.content.Username, c.content.Password))
	}

	// set accept content
	switch {
	case len(c.content.AcceptContentTypes) > 0:
		r.SetHeader("Accept", c.content.AcceptContentTypes)
	case len(c.content.ContentType) > 0:
		r.SetHeader("Accept", c.content.ContentType+", */*")
	}

	return r
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// NewRequestWithClient creates a Request with an embedded RESTClient for use in test scenarios.
func NewRequestWithClient(base *url.URL, versionedAPIPath string,
	content ClientContentConfig, client *gorequest.SuperAgent) *Request {
	return NewRequest(&RESTClient{
		base:             base,
		versionedAPIPath: versionedAPIPath,
		content:          content,
		Client:           client,
	})
}

// Verb sets the verb this request will use.
func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

// Prefix adds segments to the relative beginning to the request path. These
// items will be placed before the optional Namespace, Resource, or Name sections.
// Setting AbsPath will clear any previously set Prefix segments.
func (r *Request) Prefix(segments ...string) *Request {
	if r.err != nil {
		return r
	}

	r.pathPrefix = path.Join(r.pathPrefix, path.Join(segments...))

	return r
}

// Suffix appends segments to the end of the path. These items will be placed after the prefix and optional
// Namespace, Resource, or Name sections.
func (r *Request) Suffix(segments ...string) *Request {
	if r.err != nil {
		return r
	}

	r.subpath = path.Join(r.subpath, path.Join(segments...))

	return r
}

// Resource sets the resource to access (<resource>/[ns/<namespace>/]<name>).
func (r *Request) Resource(resource string) *Request {
	if r.err != nil {
		return r
	}

	if len(r.resource) != 0 {
		r.err = fmt.Errorf("resource already set to %q, cannot change to %q", r.resource, resource)
		return r
	}

	if msgs := IsValidPathSegmentName(resource); len(msgs) != 0 {
		r.err = fmt.Errorf("invalid resource %q: %v", resource, msgs)
		return r
	}

	r.resource = resource

	return r
}

// SubResource sets a sub-resource path which can be multiple segments after the resource
// name but before the suffix.
func (r *Request) SubResource(subresources ...string) *Request {
	if r.err != nil {
		return r
	}

	subresource := path.Join(subresources...)

	if len(r.subresource) != 0 {
		r.err = fmt.Errorf("subresource already set to %q, cannot change to %q", r.resource, subresource)
		return r
	}

	for _, s := range subresources {
		if msgs := IsValidPathSegmentName(s); len(msgs) != 0 {
			r.err = fmt.Errorf("invalid subresource %q: %v", s, msgs)
			return r
		}
	}

	r.subresource = subresource

	return r
}

// Name sets the name of a resource to access (<resource>/[ns/<namespace>/]<name>).
func (r *Request) Name(resourceName string) *Request {
	if r.err != nil {
		return r
	}

	if len(resourceName) == 0 {
		r.err = fmt.Errorf("resource name may not be empty")
		return r
	}

	if len(r.resourceName) != 0 {
		r.err = fmt.Errorf("resource name already set to %q, cannot change to %q", r.resourceName, resourceName)
		return r
	}

	if msgs := IsValidPathSegmentName(resourceName); len(msgs) != 0 {
		r.err = fmt.Errorf("invalid resource name %q: %v", resourceName, msgs)
		return r
	}

	r.resourceName = resourceName

	return r
}

// AbsPath overwrites an existing path with the segments provided. Trailing slashes are preserved
// when a single segment is passed.
func (r *Request) AbsPath(segments ...string) *Request {
	if r.err != nil {
		return r
	}

	r.pathPrefix = path.Join(r.c.base.Path, path.Join(segments...))

	if len(segments) == 1 && (len(r.c.base.Path) > 1 || len(segments[0]) > 1) && strings.HasSuffix(segments[0], "/") {
		// preserve any trailing slashes for legacy behavior
		r.pathPrefix += "/"
	}

	return r
}

// RequestURI overwrites existing path and parameters with the value of the provided server relative
// URI.
func (r *Request) RequestURI(uri string) *Request {
	if r.err != nil {
		return r
	}

	locator, err := url.Parse(uri)
	if err != nil {
		r.err = err
		return r
	}

	r.pathPrefix = locator.Path

	if len(locator.Query()) > 0 {
		if r.params == nil {
			r.params = make(url.Values)
		}

		for k, v := range locator.Query() {
			r.params[k] = v
		}
	}

	return r
}

// Param creates a query parameter with the given string value.
func (r *Request) Param(paramName, s string) *Request {
	if r.err != nil {
		return r
	}

	return r.setParam(paramName, s)
}

// VersionedParams will take the provided object, serialize it to a map[string][]string using the
// implicit RESTClient API version and the default parameter codec, and then add those as parameters
// to the request. Use this to provide versioned query parameters from client libraries.
// VersionedParams will not write query parameters that have omitempty set and are empty. If a
// parameter has already been set it is appended to (Params and VersionedParams are additive).
func (r *Request) VersionedParams(v interface{}) *Request {
	if r.err != nil {
		return r
	}

	r.c.Client.Query(v)

	return r
}

func (r *Request) setParam(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}

	r.params[paramName] = append(r.params[paramName], value)

	return r
}

// SetHeader set header for a http request.
func (r *Request) SetHeader(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}

	r.headers.Del(key)

	for _, value := range values {
		r.headers.Add(key, value)
	}

	return r
}

// Timeout makes the request use the given duration as an overall timeout for the
// request. Additionally, if set passes the value as "timeout" parameter in URL.
func (r *Request) Timeout(d time.Duration) *Request {
	if r.err != nil {
		return r
	}

	r.timeout = d

	return r
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	p := r.pathPrefix
	if len(r.resource) != 0 {
		p = path.Join(p, strings.ToLower(r.resource))
	}
	// Join trims trailing slashes, so preserve r.pathPrefix's trailing slash
	// for backwards compatibility if nothing was changed
	if len(r.resourceName) != 0 || len(r.subpath) != 0 || len(r.subresource) != 0 {
		p = path.Join(p, r.resourceName, r.subresource, r.subpath)
	}

	finalURL := &url.URL{}
	if r.c.base != nil {
		*finalURL = *r.c.base
	}

	finalURL.Path = p

	query := url.Values{}

	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	// timeout is handled specially here.
	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}

	finalURL.RawQuery = query.Encode()

	return finalURL
}

// Body makes the request use obj as the body. Optional.
func (r *Request) Body(obj interface{}) *Request {
	if v := reflect.ValueOf(obj); v.Kind() == reflect.Struct {
		r.SetHeader("Content-Type", r.c.content.ContentType)
	}

	r.body = obj

	return r
}

// Do formats and executes the request. Returns a Result object for easy response processing.
func (r *Request) Do(ctx context.Context) Result {
	client := r.c.Client
	client.Header = r.headers

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)

		defer cancel()
	}

	client.WithContext(ctx)

	resp, body, errs := client.CustomMethod(r.verb, r.URL().String()).Send(r.body).EndBytes()
	if err := combineErr(resp, body, errs); err != nil {
		return Result{
			response: &resp,
			err:      err,
			body:     body,
		}
	}

	decoder, err := r.c.content.Negotiator.Decoder()
	if err != nil {
		return Result{
			response: &resp,
			err:      err,
			body:     body,
			decoder:  decoder,
		}
	}

	return Result{
		response: &resp,
		body:     body,
		decoder:  decoder,
	}
}

// Result contains the result of calling Request.Do().
type Result struct {
	response *gorequest.Response
	err      error
	body     []byte
	decoder  runtime.Decoder
}

// Raw returns the raw result.
func (r Result) Raw() ([]byte, error) {
	return r.body, r.err
}

// Into stores the result into obj, if possible. If obj is nil it is ignored.
func (r Result) Into(v interface{}) error {
	if r.err != nil {
		return r.Error()
	}

	if r.decoder == nil {
		return fmt.Errorf("serializer doesn't exist")
	}

	if err := r.decoder.Decode(r.body, &v); err != nil {
		return err
	}

	return nil
}

// Error implements the error interface.
func (r Result) Error() error {
	return r.err
}

func combineErr(resp gorequest.Response, body []byte, errs []error) error {
	var e, sep string

	if len(errs) > 0 {
		for _, err := range errs {
			e = sep + err.Error()
			sep = "\n"
		}

		return errors.New(e)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil
}

// NameMayNotBe specifies strings that cannot be used as names specified as
// path segments (like the REST API or etcd store).
var NameMayNotBe = []string{".", ".."}

// NameMayNotContain specifies substrings that cannot be used in names specified
// as path segments (like the REST API or etcd store).
var NameMayNotContain = []string{"/", "%"}

// IsValidPathSegmentName validates the name can be safely encoded as a path segment.
func IsValidPathSegmentName(name string) []string {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return []string{fmt.Sprintf(`may not be '%s'`, illegalName)}
		}
	}

	var errors []string

	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}

// IsValidPathSegmentPrefix validates the name can be used as a prefix for a name
// which will be encoded as a path segment. It does not check for exact matches
// with disallowed names, since an arbitrary suffix might make the name valid.
func IsValidPathSegmentPrefix(name string) []string {
	var errors []string

	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}

// ValidatePathSegmentName validates the name can be safely encoded as a path segment.
func ValidatePathSegmentName(name string, prefix bool) []string {
	if prefix {
		return IsValidPathSegmentPrefix(name)
	}

	return IsValidPathSegmentName(name)
}
