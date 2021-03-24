// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package clientcmd

import (
	"net/url"
	"time"

	restclient "github.com/marmotedu/marmotedu-sdk-go/rest"
)

// Server contains information about how to communicate with a iam api server.
type Server struct {
	LocationOfOrigin string
	Timeout          time.Duration `yaml:"timeout,omitempty"                    mapstructure:"timeout,omitempty"`
	MaxRetries       int           `yaml:"max-retries,omitempty"                mapstructure:"max-retries,omitempty"`
	RetryInterval    time.Duration `yaml:"retry-interval,omitempty"             mapstructure:"retry-interval,omitempty"`
	Address          string        `yaml:"address,omitempty"                    mapstructure:"address,omitempty"`
	// TLSServerName is used to check server certificate. If TLSServerName is empty, the hostname used to contact the
	// server is used.
	// +optional
	TLSServerName string `yaml:"tls-server-name,omitempty"            mapstructure:"tls-server-name,omitempty"`
	// InsecureSkipTLSVerify skips the validity check for the server's certificate. This will make your HTTPS
	// connections insecure.
	// +optional
	InsecureSkipTLSVerify bool `yaml:"insecure-skip-tls-verify,omitempty"   mapstructure:"insecure-skip-tls-verify,omitempty"`
	// CertificateAuthority is the path to a cert file for the certificate authority.
	// +optional
	CertificateAuthority string `yaml:"certificate-authority,omitempty"      mapstructure:"certificate-authority,omitempty"`
	// CertificateAuthorityData contains PEM-encoded certificate authority certificates.
	// Overrides CertificateAuthority
	// +optional
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty" mapstructure:"certificate-authority-data,omitempty"`
}

// AuthInfo contains information that describes identity information.
// This is use to tell the iam cluster who you are.
type AuthInfo struct {
	LocationOfOrigin  string
	ClientCertificate string `yaml:"client-certificate,omitempty"      mapstructure:"client-certificate,omitempty"`
	// ClientCertificateData contains PEM-encoded data from a client cert file for TLS. Overrides ClientCertificate
	// +optional
	ClientCertificateData string `yaml:"client-certificate-data,omitempty" mapstructure:"client-certificate-data,omitempty"`
	// ClientKey is the path to a client key file for TLS.
	// +optional
	ClientKey string `yaml:"client-key,omitempty"              mapstructure:"client-key,omitempty"`
	// ClientKeyData contains PEM-encoded data from a client key file for TLS. Overrides ClientKey
	// +optional
	ClientKeyData string `yaml:"client-key-data,omitempty"         mapstructure:"client-key-data,omitempty"`
	// Token is the bearer token for authentication to the iam cluster.
	// +optional
	Token string `yaml:"token,omitempty"                   mapstructure:"token,omitempty"`

	Username string `yaml:"username,omitempty" mapstructure:"username,omitempty"`
	Password string `yaml:"password,omitempty" mapstructure:"password,omitempty"`

	SecretID  string `yaml:"secret-id,omitempty"  mapstructure:"secret-id,omitempty"`
	SecretKey string `yaml:"secret-key,omitempty" mapstructure:"secret-key,omitempty"`
}

// Config defines a config struct used by marmotedu-sdk-go.
type Config struct {
	APIVersion string    `yaml:"apiVersion,omitempty" mapstructure:"apiVersion,omitempty"`
	AuthInfo   *AuthInfo `yaml:"user,omitempty"       mapstructure:"user,omitempty"`
	Server     *Server   `yaml:"server,omitempty"     mapstructure:"server,omitempty"`
}

// NewConfig is a convenience function that returns a new Config object with non-nil maps.
func NewConfig() *Config {
	return &Config{
		Server:   &Server{},
		AuthInfo: &AuthInfo{},
	}
}

// ClientConfig is used to make it easy to get an api server client.
type ClientConfig interface {
	// ClientConfig returns a complete client config
	ClientConfig() (*restclient.Config, error)
}

// DirectClientConfig wrap for Config.
type DirectClientConfig struct {
	config Config
}

// NewClientConfigFromConfig takes your Config and gives you back a ClientConfig.
func NewClientConfigFromConfig(config *Config) ClientConfig {
	return &DirectClientConfig{*config}
}

// NewClientConfigFromBytes takes your iamconfig and gives you back a ClientConfig.
func NewClientConfigFromBytes(configBytes []byte) (ClientConfig, error) {
	config, err := Load(configBytes)
	if err != nil {
		return nil, err
	}

	return &DirectClientConfig{*config}, nil
}

// RESTConfigFromIAMConfig is a convenience method to give back a restconfig from your iamconfig bytes.
// For programmatic access, this is what you want 80% of the time.
func RESTConfigFromIAMConfig(configBytes []byte) (*restclient.Config, error) {
	clientConfig, err := NewClientConfigFromBytes(configBytes)
	if err != nil {
		return nil, err
	}

	return clientConfig.ClientConfig()
}

// ClientConfig implements ClientConfig.
func (config *DirectClientConfig) ClientConfig() (*restclient.Config, error) {
	user := config.getAuthInfo()
	server := config.getServer()

	if err := config.ConfirmUsable(); err != nil {
		return nil, err
	}

	clientConfig := &restclient.Config{
		BearerToken:   user.Token,
		Username:      user.Username,
		Password:      user.Password,
		SecretID:      user.SecretID,
		SecretKey:     user.SecretKey,
		Host:          server.Address,
		Timeout:       server.Timeout,
		MaxRetries:    server.MaxRetries,
		RetryInterval: server.RetryInterval,
		TLSClientConfig: restclient.TLSClientConfig{
			Insecure:   server.InsecureSkipTLSVerify,
			ServerName: server.TLSServerName,
			CertFile:   user.ClientCertificate,
			KeyFile:    user.ClientKey,
			CertData:   []byte(user.ClientCertificateData),
			KeyData:    []byte(user.ClientKeyData),
			CAFile:     server.CertificateAuthority,
			CAData:     []byte(server.CertificateAuthorityData),
			// NextProtos []string
		},
	}

	if u, err := url.ParseRequestURI(clientConfig.Host); err == nil && u.Opaque == "" && len(u.Path) > 1 {
		u.RawQuery = ""
		u.Fragment = ""
		clientConfig.Host = u.String()
	}

	return clientConfig, nil
}

// ConfirmUsable looks a particular context and determines if that particular part of
// the config is useable.  There might still be errors in the config, but no errors in the
// sections requested or referenced.  It does not return early so that it can find as many errors as possible.
func (config *DirectClientConfig) ConfirmUsable() error {
	validationErrors := make([]error, 0)

	authInfo := config.getAuthInfo()
	validationErrors = append(validationErrors, validateAuthInfo(authInfo)...)
	server := config.getServer()
	validationErrors = append(validationErrors, validateServerInfo(server)...)
	// when direct client config is specified, and our only error is that no server is defined, we should
	// return a standard "no config" error
	if len(validationErrors) == 1 && validationErrors[0] == ErrEmptyServer {
		return newErrConfigurationInvalid([]error{ErrEmptyConfig})
	}

	return newErrConfigurationInvalid(validationErrors)
}

// getAuthInfo returns the clientcmdapi.AuthInfo, or an error if a required auth info is not found.
func (config *DirectClientConfig) getAuthInfo() AuthInfo {
	return *config.config.AuthInfo
}

// getServer returns the clientcmdapi.Cluster, or an error if a required cluster is not found.
func (config *DirectClientConfig) getServer() Server {
	return *config.config.Server
}

// BuildConfigFromFlags is a helper function that builds configs from a master
// url or a iamconfig filepath. These are passed in as command line flags for cluster
// components. Warnings should reflect this usage. If neither masterUrl or iamconfigPath
// are passed in we fallback to inClusterConfig. If inClusterConfig fails, we fallback
// to the default config.
func BuildConfigFromFlags(serverURL, iamconfigPath string) (*restclient.Config, error) {
	config, err := LoadFromFile(iamconfigPath)
	if err != nil {
		return nil, err
	}

	if len(serverURL) > 0 {
		config.Server.Address = serverURL
	}

	directClientConfig := &DirectClientConfig{*config}

	return directClientConfig.ClientConfig()
}
