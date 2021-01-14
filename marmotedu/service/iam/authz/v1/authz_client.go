// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	v1 "github.com/marmotedu/api/authz/v1"
	"github.com/marmotedu/component-base/pkg/runtime"

	"github.com/marmotedu/marmotedu-sdk-go/rest"
)

// AuthzV1Interface has methods to work with iam resources.
type AuthzV1Interface interface {
	RESTClient() rest.Interface
	AuthzGetter
}

// AuthzV1Client is used to interact with features provided by the group.
type AuthzV1Client struct {
	restClient rest.Interface
}

// Authz create and return authz rest client.
func (c *AuthzV1Client) Authz() AuthzInterface {
	return newAuthz(c)
}

// NewForConfig creates a new AuthzV1Client for the given config.
func NewForConfig(c *rest.Config) (*AuthzV1Client, error) {
	config := *c
	setConfigDefaults(&config)

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &AuthzV1Client{client}, nil
}

// NewForConfigOrDie creates a new AuthzV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *AuthzV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}

	return client
}

// New creates a new AuthzV1Client for the given RESTClient.
func New(c rest.Interface) *AuthzV1Client {
	return &AuthzV1Client{c}
}

func setConfigDefaults(config *rest.Config) {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = ""
	config.Negotiator = runtime.NewSimpleClientNegotiator()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultUserAgent()
	}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *AuthzV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}

	return c.restClient
}
