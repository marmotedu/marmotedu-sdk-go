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

// SecretsGetter has a method to return a SecretInterface.
// A group's client should implement this interface.
type SecretsGetter interface {
	Secrets() SecretInterface
}

// SecretInterface has methods to work with Secret resources.
type SecretInterface interface {
	Create(ctx context.Context, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error)
	Update(ctx context.Context, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Secret, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.SecretList, error)
	SecretExpansion
}

// secrets implements SecretInterface.
type secrets struct {
	client rest.Interface
}

// newSecrets returns a Secrets.
func newSecrets(c *APIV1Client) *secrets {
	return &secrets{
		client: c.RESTClient(),
	}
}

// Get takes name of the secret, and returns the corresponding secret object, and an error if there is any.
func (c *secrets) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Secret, err error) {
	result = &v1.Secret{}
	err = c.client.Get().
		Resource("secrets").
		Name(name).
		VersionedParams(options).
		Do(ctx).
		Into(result)

	return
}

// List takes label and field selectors, and returns the list of Secrets that match those selectors.
func (c *secrets) List(ctx context.Context, opts metav1.ListOptions) (result *v1.SecretList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1.SecretList{}
	err = c.client.Get().
		Resource("secrets").
		VersionedParams(opts).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}

// Create takes the representation of a secret and creates it.
// Returns the server's representation of the secret, and an error, if there is any.
func (c *secrets) Create(ctx context.Context, secret *v1.Secret,
	opts metav1.CreateOptions) (result *v1.Secret, err error) {
	result = &v1.Secret{}
	err = c.client.Post().
		Resource("secrets").
		VersionedParams(opts).
		Body(secret).
		Do(ctx).
		Into(result)

	return
}

// Update takes the representation of a secret and updates it.
// Returns the server's representation of the secret, and an error, if there is any.
func (c *secrets) Update(ctx context.Context, secret *v1.Secret,
	opts metav1.UpdateOptions) (result *v1.Secret, err error) {
	result = &v1.Secret{}
	err = c.client.Put().
		Resource("secrets").
		Name(secret.Name).
		VersionedParams(opts).
		Body(secret).
		Do(ctx).
		Into(result)

	return
}

func (c *secrets) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("secrets").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *secrets) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}

	return c.client.Delete().
		Resource("secrets").
		VersionedParams(listOpts).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}
