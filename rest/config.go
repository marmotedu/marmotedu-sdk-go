// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rest

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	gruntime "runtime"
	"strings"
	"time"

	"github.com/marmotedu/component-base/pkg/runtime"
	"github.com/marmotedu/component-base/pkg/scheme"

	"github.com/marmotedu/marmotedu-sdk-go/pkg/version"
	"github.com/marmotedu/marmotedu-sdk-go/third_party/forked/gorequest"
)

// Config holds the common attributes that can be passed to a IAM client on
// initialization.
type Config struct {
	Host    string
	APIPath string
	ContentConfig

	// Server requires Basic authentication
	Username string
	Password string

	SecretID  string
	SecretKey string

	// Server requires Bearer authentication. This client will not attempt to use
	// refresh tokens for an OAuth2 flow.
	// TODO: demonstrate an OAuth2 compatible client.
	BearerToken string

	// Path to a file containing a BearerToken.
	// If set, the contents are periodically read.
	// The last successfully read value takes precedence over BearerToken.
	BearerTokenFile string

	// TLSClientConfig contains settings to enable transport layer security
	TLSClientConfig

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string
	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// ContentConfig defines config for content.
type ContentConfig struct {
	ServiceName        string
	AcceptContentTypes string
	ContentType        string
	GroupVersion       *scheme.GroupVersion
	Negotiator         runtime.ClientNegotiator
}

type sanitizedConfig *Config

// GoString implements fmt.GoStringer and sanitizes sensitive fields of Config
// to prevent accidental leaking via logs.
func (c *Config) GoString() string {
	return c.String()
}

// String implements fmt.Stringer and sanitizes sensitive fields of Config to
// prevent accidental leaking via logs.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}

	cc := sanitizedConfig(CopyConfig(c))
	// Explicitly mark non-empty credential fields as redacted.
	if cc.Password != "" {
		cc.Password = "--- REDACTED ---"
	}

	if cc.BearerToken != "" {
		cc.BearerToken = "--- REDACTED ---"
	}

	if cc.SecretKey != "" {
		cc.SecretKey = "--- REDACTED ---"
	}

	return fmt.Sprintf("%#v", cc)
}

// TLSClientConfig contains settings to enable transport layer security.
type TLSClientConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool
	// ServerName is passed to the server for SNI and is used in the client to check server
	// ceritificates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	ServerName string

	// Server requires TLS client certificate authentication
	CertFile string
	// Server requires TLS client certificate authentication
	KeyFile string
	// Trusted root certificates for server
	CAFile string

	// CertData holds PEM-encoded bytes (typically read from a client certificate file).
	// CertData takes precedence over CertFile
	CertData []byte
	// KeyData holds PEM-encoded bytes (typically read from a client certificate key file).
	// KeyData takes precedence over KeyFile
	KeyData []byte
	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	CAData []byte

	// NextProtos is a list of supported application level protocols, in order of preference.
	// Used to populate tls.Config.NextProtos.
	// To indicate to the server http/1.1 is preferred over http/2, set to ["http/1.1", "h2"] (though the server is free
	// to ignore that preference).
	// To use only http/1.1, set to ["http/1.1"].
	NextProtos []string
}

var (
	_ fmt.Stringer   = TLSClientConfig{}
	_ fmt.GoStringer = TLSClientConfig{}
)

type sanitizedTLSClientConfig TLSClientConfig

// GoString implements fmt.GoStringer and sanitizes sensitive fields of
// TLSClientConfig to prevent accidental leaking via logs.
func (c TLSClientConfig) GoString() string {
	return c.String()
}

// String implements fmt.Stringer and sanitizes sensitive fields of
// TLSClientConfig to prevent accidental leaking via logs.
func (c TLSClientConfig) String() string {
	// nolint: gosimple // no need
	cc := sanitizedTLSClientConfig{
		Insecure:   c.Insecure,
		ServerName: c.ServerName,
		CertFile:   c.CertFile,
		KeyFile:    c.KeyFile,
		CAFile:     c.CAFile,
		CertData:   c.CertData,
		KeyData:    c.KeyData,
		CAData:     c.CAData,
		NextProtos: c.NextProtos,
	}
	// Explicitly mark non-empty credential fields as redacted.
	if len(cc.CertData) != 0 {
		cc.CertData = []byte("--- TRUNCATED ---")
	}

	if len(cc.KeyData) != 0 {
		cc.KeyData = []byte("--- REDACTED ---")
	}

	return fmt.Sprintf("%#v", cc)
}

// HasCA returns whether the configuration has a certificate authority or not.
func (c TLSClientConfig) HasCA() bool {
	return len(c.CAData) > 0 || len(c.CAFile) > 0
}

// HasCertAuth returns whether the configuration has certificate authentication or not.
func (c TLSClientConfig) HasCertAuth() bool {
	return (len(c.CertData) != 0 || len(c.CertFile) != 0) && (len(c.KeyData) != 0 || len(c.KeyFile) != 0)
}

// RESTClientFor returns a RESTClient that satisfies the requested attributes on a client Config
// object. Note that a RESTClient may require fields that are optional when initializing a Client.
// A RESTClient created by this method is generic - it expects to operate on an API that follows
// the IAM conventions, but may not be the IAM API.
func RESTClientFor(config *Config) (*RESTClient, error) {
	if config.GroupVersion == nil {
		return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
	}

	if config.Negotiator == nil {
		return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
	}

	baseURL, versionedAPIPath, err := defaultServerURLFor(config)
	if err != nil {
		return nil, err
	}

	// Get the TLS options for this client config
	tlsConfig, err := TLSConfigFor(config)
	if err != nil {
		return nil, err
	}

	// Only retry when get a server side error.
	client := gorequest.New().TLSClientConfig(tlsConfig).Timeout(config.Timeout).
		Retry(config.MaxRetries, config.RetryInterval, http.StatusInternalServerError)
	// NOTICE: must set DoNotClearSuperAgent to true, or the client will clean header befor http.Do
	client.DoNotClearSuperAgent = true

	var gv scheme.GroupVersion
	if config.GroupVersion != nil {
		gv = *config.GroupVersion
	}

	clientContent := ClientContentConfig{
		Username:           config.Username,
		Password:           config.Password,
		SecretID:           config.SecretID,
		SecretKey:          config.SecretKey,
		BearerToken:        config.BearerToken,
		BearerTokenFile:    config.BearerTokenFile,
		TLSClientConfig:    config.TLSClientConfig,
		AcceptContentTypes: config.AcceptContentTypes,
		ContentType:        config.ContentType,
		GroupVersion:       gv,
		Negotiator:         config.Negotiator,
	}

	return NewRESTClient(baseURL, versionedAPIPath, clientContent, client)
}

// TLSConfigFor returns a tls.Config that will provide the transport level security defined
// by the provided Config. Will return nil if no transport level security is requested.
func TLSConfigFor(c *Config) (*tls.Config, error) {
	if !(c.HasCA() || c.HasCertAuth() || c.Insecure || len(c.ServerName) > 0) {
		return nil, nil
	}

	if c.HasCA() && c.Insecure {
		return nil, fmt.Errorf("specifying a root certificates file with the insecure flag is not allowed")
	}

	if err := LoadTLSFiles(c); err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion: tls.VersionTLS12,
		//nolint: gosec
		InsecureSkipVerify: c.Insecure,
		ServerName:         c.ServerName,
		NextProtos:         c.NextProtos,
	}

	if c.HasCA() {
		tlsConfig.RootCAs = rootCertPool(c.CAData)
	}

	var staticCert *tls.Certificate
	// Treat cert as static if either key or cert was data, not a file
	if c.HasCertAuth() {
		// If key/cert were provided, verify them before setting up
		// tlsConfig.GetClientCertificate.
		cert, err := tls.X509KeyPair(c.CertData, c.KeyData)
		if err != nil {
			return nil, err
		}

		staticCert = &cert
	}

	if c.HasCertAuth() {
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			// Note: static key/cert data always take precedence over cert
			// callback.
			if staticCert != nil {
				return staticCert, nil
			}

			// Both c.TLS.CertData/KeyData were unset and GetCert didn't return
			// anything. Return an empty tls.Certificate, no client cert will
			// be sent to the server.
			return &tls.Certificate{}, nil
		}
	}

	return tlsConfig, nil
}

// rootCertPool returns nil if caData is empty.  When passed along, this will mean "use system CAs".
// When caData is not empty, it will be the ONLY information used in the CertPool.
func rootCertPool(caData []byte) *x509.CertPool {
	// What we really want is a copy of x509.systemRootsPool, but that isn't exposed.  It's difficult to build (see the
	// go
	// code for a look at the platform specific insanity), so we'll use the fact that RootCAs == nil gives us the system
	// values
	// It doesn't allow trusting either/or, but hopefully that won't be an issue
	if len(caData) == 0 {
		return nil
	}

	// if we have caData, use it
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)

	return certPool
}

// LoadTLSFiles copies the data from the CertFile, KeyFile, and CAFile fields into the CertData,
// KeyData, and CAFile fields, or returns an error. If no error is returned, all three fields are
// either populated or were empty to start.
func LoadTLSFiles(c *Config) error {
	var err error

	c.CAData, err = dataFromSliceOrFile(c.CAData, c.CAFile)
	if err != nil {
		return err
	}

	c.CertData, err = dataFromSliceOrFile(c.CertData, c.CertFile)
	if err != nil {
		return err
	}

	c.KeyData, err = dataFromSliceOrFile(c.KeyData, c.KeyFile)
	if err != nil {
		return err
	}

	return nil
}

// dataFromSliceOrFile returns data from the slice (if non-empty), or from the file,
// or an error if an error occurred reading the file.
func dataFromSliceOrFile(data []byte, file string) ([]byte, error) {
	if len(data) > 0 {
		return base64.StdEncoding.DecodeString(string(data))
	}

	if len(file) > 0 {
		fileData, err := ioutil.ReadFile(file)
		if err != nil {
			return []byte{}, err
		}

		return fileData, nil
	}

	return nil, nil
}

// SetIAMDefaults sets default values on the provided client config for accessing the
// IAM API or returns an error if any of the defaults are impossible or invalid.
func SetIAMDefaults(config *Config) error {
	if len(config.UserAgent) == 0 {
		config.UserAgent = DefaultUserAgent()
	}

	return nil
}

// adjustCommit returns sufficient significant figures of the commit's git hash.
func adjustCommit(c string) string {
	if len(c) == 0 {
		return "unknown"
	}

	if len(c) > 7 {
		return c[:7]
	}

	return c
}

// adjustVersion strips "alpha", "beta", etc. from version in form
// major.minor.patch-[alpha|beta|etc].
func adjustVersion(v string) string {
	if len(v) == 0 {
		return "unknown"
	}

	seg := strings.SplitN(v, "-", 2)

	return seg[0]
}

// adjustCommand returns the last component of the
// OS-specific command path for use in User-Agent.
func adjustCommand(p string) string {
	// Unlikely, but better than returning "".
	if len(p) == 0 {
		return "unknown"
	}

	return filepath.Base(p)
}

// buildUserAgent builds a User-Agent string from given args.
func buildUserAgent(command, version, os, arch, commit string) string {
	return fmt.Sprintf(
		"%s/%s (%s/%s) iam/%s", command, version, os, arch, commit)
}

// DefaultUserAgent returns a User-Agent string built from static global vars.
func DefaultUserAgent() string {
	return buildUserAgent(
		adjustCommand(os.Args[0]),
		adjustVersion(version.Get().GitVersion),
		gruntime.GOOS,
		gruntime.GOARCH,
		adjustCommit(version.Get().GitCommit))
}

// AddUserAgent add a http User-Agent header.
func AddUserAgent(config *Config, userAgent string) *Config {
	fullUserAgent := DefaultUserAgent() + "/" + userAgent
	config.UserAgent = fullUserAgent

	return config
}

// CopyConfig returns a copy of the given config.
func CopyConfig(config *Config) *Config {
	return &Config{
		Host:            config.Host,
		APIPath:         config.APIPath,
		ContentConfig:   config.ContentConfig,
		Username:        config.Username,
		Password:        config.Password,
		SecretID:        config.SecretID,
		SecretKey:       config.SecretKey,
		BearerToken:     config.BearerToken,
		BearerTokenFile: config.BearerTokenFile,
		TLSClientConfig: TLSClientConfig{
			Insecure:   config.TLSClientConfig.Insecure,
			ServerName: config.TLSClientConfig.ServerName,
			CertFile:   config.TLSClientConfig.CertFile,
			KeyFile:    config.TLSClientConfig.KeyFile,
			CAFile:     config.TLSClientConfig.CAFile,
			CertData:   config.TLSClientConfig.CertData,
			KeyData:    config.TLSClientConfig.KeyData,
			CAData:     config.TLSClientConfig.CAData,
			NextProtos: config.TLSClientConfig.NextProtos,
		},
		UserAgent: config.UserAgent,
		Timeout:   config.Timeout,
	}
}
