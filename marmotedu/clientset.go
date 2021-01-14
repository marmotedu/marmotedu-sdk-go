// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package marmotedu

import (
	"github.com/marmotedu/marmotedu-sdk-go/marmotedu/service/iam"
	"github.com/marmotedu/marmotedu-sdk-go/rest"
)

// Interface defines method used to return client interface used by marmotedu organization.
type Interface interface {
	Iam() iam.IamInterface
	// Tms() tms.TmsInterface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	iam *iam.IamClient
	// tms *tms.TmsClient
}

var _ Interface = &Clientset{}

// Iam retrieves the IamClient.
func (c *Clientset) Iam() iam.IamInterface {
	return c.iam
}

// Tms retrieves the TmsClient.
// func (c *Clientset) Tms() tms.TmsInterface {
//	return c.tms
// }

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c

	var cs Clientset

	var err error

	cs.iam, err = iam.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	/*
		cs.tms, err = tms.NewForConfig(&configShallowCopy)
		if err != nil {
			return nil, err
		}
	*/
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.iam = iam.NewForConfigOrDie(c)
	// cs.tms = tms.NewForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.iam = iam.New(c)
	// cs.tms = tms.New(c)
	return &cs
}
