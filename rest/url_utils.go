// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rest

import (
	"fmt"
	"net/url"
	"path"

	"github.com/marmotedu/component-base/pkg/scheme"
)

// DefaultServerURL converts a host, host:port, or URL string to the default base server API path
// to use with a Client at a given API version following the standard conventions for a
// IAM API.
func DefaultServerURL(host, apiPath string, groupVersion scheme.GroupVersion,
	defaultTLS bool) (*url.URL, string, error) {
	hostURL, err := url.Parse(host)
	if err != nil || hostURL.Scheme == "" || hostURL.Host == "" {
		requestURL := fmt.Sprintf("http://%s.marmotedu.com:8080", groupVersion.Group)
		if defaultTLS {
			requestURL = fmt.Sprintf("https://%s.marmotedu.com:8443", groupVersion.Group)
		}

		hostURL, err = url.Parse(requestURL)
		if err != nil {
			return nil, "", err
		}

		if hostURL.Path != "" && hostURL.Path != "/" {
			return nil, "", fmt.Errorf("host must be a URL or a host:port pair: %q", host)
		}
	}

	// hostURL.Path is optional; a non-empty Path is treated as a prefix that is to be applied to
	// all URIs used to access the host. this is useful when there's a proxy in front of the
	// apiserver that has relocated the apiserver endpoints, forwarding all requests from, for
	// example, /a/b/c to the apiserver. in this case the Path should be /a/b/c.
	//
	// if running without a frontend proxy (that changes the location of the apiserver), then
	// hostURL.Path should be blank.
	//
	// versionedAPIPath, a path relative to baseURL.Path, points to a versioned API base
	// versionedAPIPath = DefaultVersionedAPIPath(apiPath, groupVersion)
	versionedAPIPath := path.Join("/", apiPath, groupVersion.Version)

	return hostURL, versionedAPIPath, nil
}

// DefaultVersionedAPIPath constructs the default path for the given group version, assuming the given
// API path, following the standard conventions of the IAM API.
func DefaultVersionedAPIPath(apiPath string, groupVersion scheme.GroupVersion) string {
	versionedAPIPath := path.Join("/", apiPath)

	// Add the version to the end of the path
	if len(groupVersion.Group) > 0 {
		versionedAPIPath = path.Join(versionedAPIPath, groupVersion.Group, groupVersion.Version)
	} else {
		versionedAPIPath = path.Join(versionedAPIPath, groupVersion.Version)
	}

	return versionedAPIPath
}

// defaultServerURLFor is shared between IsConfigTransportTLS and RESTClientFor. It
// requires Host and Version to be set prior to being called.
func defaultServerURLFor(config *Config) (*url.URL, string, error) {
	// TODO: move the default to secure when the apiserver supports TLS by default
	// config.Insecure is taken to mean "I want HTTPS but don't bother checking the certs against a CA."
	hasCA := len(config.CAFile) != 0 || len(config.CAData) != 0
	hasCert := len(config.CertFile) != 0 || len(config.CertData) != 0
	defaultTLS := hasCA || hasCert || config.Insecure

	if config.GroupVersion != nil {
		return DefaultServerURL(config.Host, config.APIPath, *config.GroupVersion, defaultTLS)
	}

	return DefaultServerURL(config.Host, config.APIPath, scheme.GroupVersion{}, defaultTLS)
}
