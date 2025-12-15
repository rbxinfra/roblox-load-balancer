package types

// ServiceConfig represents the configuration
// for a service.
type ServiceConfig struct {
	// Protocol is the protocol to use when
	// requesting the server.
	//
	// For https, you can optionally add a CA bundle
	// to verify the hosts against, or just disable
	// TLS verification checks.
	//
	// One of: http, https, h2c (for insecure HTTP2)
	// Default: http
	Protocol string

	// Enable determines if this is enabled or not
	// always true
	Enable bool

	// Fe is the frontend configuration
	Fe *FrontendConfiguration

	// Be is the backend configuration
	Be *BackendConfiguration
}

// Hash computes a hash of the ServiceConfig
func (sc *ServiceConfig) Hash() uint64 {
	var hash uint64 = 17

	hash = hash*31 + uint64(len(sc.Protocol))
	for i := 0; i < len(sc.Protocol); i++ {
		hash = hash*31 + uint64(sc.Protocol[i])
	}

	if sc.Enable {
		hash = hash*31 + 1
	}

	if sc.Fe != nil {
		hash = hash*31 + sc.Fe.Hash()
	}

	if sc.Be != nil {
		hash = hash*31 + sc.Be.Hash()
	}

	return hash
}
