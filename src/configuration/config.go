package configuration

// Config represents the configuration for
// the HAProxy config builder.
type Config struct {
	// Prefix is the prefix within the labels to use
	// for both label parsing and filtering.
	//
	// Defaults to "haproxy"
	Prefix string `json:"prefix" yaml:"prefix" toml:"prefix"`

	// TemplateFilePath is the path to the templated file
	// that will be used to build the final HAProxy configuration.
	//
	// This is required.
	TemplateFilePath string `json:"templateFilePath" yaml:"template_file_path" toml:"template_file_path"`

	// OutputFilePath is the path to the file where the built
	// templated file is written to.
	//
	// Defaults to /usr/local/etc/haproxy/haproxy.cfg
	OutputFilePath string `json:"outputFilePath" yaml:"output_file_path" toml:"output_file_path"`

	// TLSBundleFilePath is the path to the TLS bundle used
	// to verify TLS requests to backends.
	TLSBundleFilePath string `json:"tlsBundleFilePath" yaml:"tls_bundle_file_path" toml:"tls_bundle_file_path"`

	// Entrypoints represents the configuration on backends
	// per entrypoint.
	Entrypoints map[string]*EntrypointConfig `json:"entryPoints" yaml:"entrypoints" toml:"entrypoints"`

	// HealthChecks represents the individual health check config
	// for a service, or a default configuration for all services.
	HealthChecks map[string]*HealthCheckConfig `json:"healthChecks" yaml:"health_checks" toml:"health_checks"`

	ServersConfig *ServersConfig `json:"servers" yaml:"servers" toml:"servers"`

	// Consul represents the Consul configuration options.
	Consul *ConsulConfig `json:"consul" yaml:"consul" toml:"consul"`

	// HAProxy represents the HAProxy configuration options.
	HAProxy *HAProxyConfig `json:"haproxy" yaml:"haproxy" toml:"haproxy"`
}
