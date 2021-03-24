// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	"context"

	authzv1 "github.com/marmotedu/api/authz/v1"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/ory/ladon"

	rest "github.com/marmotedu/marmotedu-sdk-go/rest"
)

// AuthzGetter has a method to return a AuthzInterface.
// A group's client should implement this interface.
type AuthzGetter interface {
	Authz() AuthzInterface
}

// AuthzInterface has methods to work with Authz resources.
type AuthzInterface interface {
	Authorize(ctx context.Context, request *ladon.Request, opts metav1.AuthorizeOptions) (*authzv1.Response, error)
	AuthzExpansion
}

// authz implements AuthzInterface.
type authz struct {
	client rest.Interface
}

// newAuthz returns a Authz.
func newAuthz(c *AuthzV1Client) *authz {
	return &authz{
		client: c.RESTClient(),
	}
}

// Get takes name of the secret, and returns the corresponding secret object, and an error if there is any.
func (c *authz) Authorize(ctx context.Context, request *ladon.Request,
	opts metav1.AuthorizeOptions) (result *authzv1.Response, err error) {
	result = &authzv1.Response{}
	err = c.client.Post().
		Resource("authz").
		VersionedParams(opts).
		Body(request).
		Do(ctx).
		Into(result)

	return
}
