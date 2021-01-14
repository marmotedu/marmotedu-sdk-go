// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	"context"
	"time"

	v1 "github.com/marmotedu/api/apiserver/v1"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"

	rest "github.com/marmotedu/marmotedu-sdk-go/rest"
)

// UsersGetter has a method to return a UserInterface.
// A group's client should implement this interface.
type UsersGetter interface {
	Users() UserInterface
}

// UserInterface has methods to work with User resources.
type UserInterface interface {
	Create(ctx context.Context, user *v1.User, opts metav1.CreateOptions) (*v1.User, error)
	Update(ctx context.Context, user *v1.User, opts metav1.UpdateOptions) (*v1.User, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.User, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.UserList, error)
	UserExpansion
}

// users implements UserInterface.
type users struct {
	client rest.Interface
}

// newUsers returns a Users.
func newUsers(c *APIV1Client) *users {
	return &users{
		client: c.RESTClient(),
	}
}

// Get takes name of the user, and returns the corresponding user object, and an error if there is any.
func (c *users) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.User, err error) {
	result = &v1.User{}
	err = c.client.Get().
		Resource("users").
		Name(name).
		VersionedParams(options).
		Do(ctx).
		Into(result)

	return
}

// List takes label and field selectors, and returns the list of Users that match those selectors.
func (c *users) List(ctx context.Context, opts metav1.ListOptions) (result *v1.UserList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1.UserList{}
	err = c.client.Get().
		Resource("users").
		VersionedParams(opts).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}

// Create takes the representation of a user and creates it.
// Returns the server's representation of the user, and an error, if there is any.
func (c *users) Create(ctx context.Context, user *v1.User, opts metav1.CreateOptions) (result *v1.User, err error) {
	result = &v1.User{}
	err = c.client.Post().
		Resource("users").
		VersionedParams(opts).
		Body(user).
		Do(ctx).
		Into(result)

	return
}

// Update takes the representation of a user and updates it.
// Returns the server's representation of the user, and an error, if there is any.
func (c *users) Update(ctx context.Context, user *v1.User, opts metav1.UpdateOptions) (result *v1.User, err error) {
	result = &v1.User{}
	err = c.client.Put().
		Resource("users").
		Name(user.Name).
		VersionedParams(opts).
		Body(user).
		Do(ctx).
		Into(result)

	return
}

func (c *users) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("users").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *users) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}

	return c.client.Delete().
		Resource("users").
		VersionedParams(listOpts).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}
