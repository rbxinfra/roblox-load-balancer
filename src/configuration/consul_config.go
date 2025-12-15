package configuration

import "time"

// ConsulHttpBasicAuth is the basic HTTP authentication
// details.
type ConsulHttpBasicAuth struct {
	// Username to use for HTTP Basic Authentication
	Username string `json:"username" yaml:"username" toml:"username"`

	// Password to use for HTTP Basic Authentication
	Password string `json:"password" yaml:"password" toml:"password"`
}

// ConsulTLSConfig represents the Consul API TLS configuration
type ConsulTLSConfig struct {
	// Address is the optional address of the Consul server. The port, if any
	// will be removed from here and this will be set to the ServerName of the
	// resulting config.
	Address string `json:"address" yaml:"address" toml:"address"`

	// CAFile is the optional path to the CA certificate used for Consul
	// communication, defaults to the system bundle if not specified.
	CAFile string `json:"caFile" yaml:"ca_file" toml:"ca_file"`

	// CAPath is the optional path to a directory of CA certificates to use for
	// Consul communication, defaults to the system bundle if not specified.
	CAPath string `json:"caPath" yaml:"ca_path" toml:"ca_path"` 

	// CAPem is the optional PEM-encoded CA certificate used for Consul
	// communication, defaults to the system bundle if not specified.
	CAPem []byte `json:"caPem" yaml:"ca_pem" toml:"ca_pem"`

	// CertFile is the optional path to the certificate for Consul
	// communication. If this is set then you need to also set KeyFile.
	CertFile string `json:"certFile" yaml:"cert_file" toml:"cert_file"`

	// CertPEM is the optional PEM-encoded certificate for Consul
	// communication. If this is set then you need to also set KeyPEM.
	CertPEM []byte `json:"certPem" yaml:"cert_pem" toml:"cert_pem"`

	// KeyFile is the optional path to the private key for Consul communication.
	// If this is set then you need to also set CertFile.
	KeyFile string `json:"keyFile" yaml:"key_file" toml:"key_file"`

	// KeyPEM is the optional PEM-encoded private key for Consul communication.
	// If this is set then you need to also set CertPEM.
	KeyPEM []byte `json:"keyPem" yaml:"key_pem" toml:"key_pem"`

	// InsecureSkipVerify if set to true will disable TLS host verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecure_skip_verify" toml:"insecure_skip_verify"`
}

// ConsulConfig represents the configuration
// for Consul service discovery.
// Pretty much just config options for the API client.
type ConsulConfig struct {
	// Address is the address of the Consul server.
	Address string `json:"address" yaml:"address" toml:"address"`

	// Scheme is the URI scheme for the Consul server.
	Scheme string `json:"scheme" yaml:"scheme" toml:"scheme"`

	// PathPrefix for URIs for when Consul is behind an API gateway (reverse
	// proxy). The API gatewat must strip off the PathPrefix before passing
	// the request onto Consul.
	PathPrefix string `json:"pathPrefix" yaml:"path_prefix" toml:"path_prefix"`

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string `json:"datacenter" yaml:"datacenter" toml:"datacenter"`

	// HttpAuth is the auth info to use for HTTP access.
	HttpAuth *ConsulHttpBasicAuth `json:"httpAuth" yaml:"http_auth" toml:"http_auth"`
	
	// WaitTime limits how long a watch will block. If not provided,
	// the agent default values will be used.
	WaitTime time.Duration `json:"waitTime" yaml:"wait_time" toml:"wait_time"`

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	Token string `json:"token" yaml:"token" toml:"token"`

	// TokenFile is a file containing the current token tro use for this client.
	// If provided it is read once at startup and never again.
	TokenFile string `json:"tokenFile" yaml:"token_file" toml:"token_file"`

	// Namespace is the name of the namespace to send along for the request
	// when no other Namespace is present in the QueryOptions.
	Namespace string `json:"namespace" yaml:"namespace" toml:"namespace"`

	// Partition is the name of the partition to send along for the request
	// when no other Paritition is present in the QueryOptions.
	Partition string `json:"partition" yaml:"partition" toml:"partition"`

	// TLSConfig is the configuration for the TLS client.
	TLSConfig *ConsulTLSConfig `json:"tlsConfig" yaml:"tls_config" toml:"tls_config"`
}
