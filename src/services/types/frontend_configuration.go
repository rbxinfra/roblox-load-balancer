package types

// FrontendConfiguration is anything related to a frontend for
// a service.
// Usually HTTP rules.
type FrontendConfiguration struct {
	// Fqdn is a list of hosts to use to resolve
	// a backend.
	Fqdn []string

	// BlockedPaths is a list of directly blocked paths.
	// The difference between FE and BE blocked paths
	// is that if any of these paths hit then this will
	// be treated as an unknown backend vs 403.
	BlockedPaths []string

	// BlockedPaths_Beg is a list of blocked path prefixes.
	// The difference between FE and BE blocked paths
	// is that if any of these paths hit then this will
	// be treated as an unknown backend vs 403.
	BlockedPaths_Beg []string

	// EntryPoints is the list of entrypoints that this
	// service can be called from.
	EntryPoints []string

	// PathPrefix is the prefix to use when routing
	// to the backend.
	//
	// This will be stripped on the backend.
	PathPrefix string
}

// Hash computes a hash of the FrontendConfiguration
func (fc *FrontendConfiguration) Hash() uint64 {
	var hash uint64 = 17

	hash = hash*31 + uint64(len(fc.Fqdn))
	for _, fqdn := range fc.Fqdn {
		hash = hash*31 + uint64(len(fqdn))
		for i := 0; i < len(fqdn); i++ {
			hash = hash*31 + uint64(fqdn[i])
		}
	}

	hash = hash*31 + uint64(len(fc.BlockedPaths))
	for _, path := range fc.BlockedPaths {
		hash = hash*31 + uint64(len(path))
		for i := 0; i < len(path); i++ {
			hash = hash*31 + uint64(path[i])
		}
	}

	hash = hash*31 + uint64(len(fc.BlockedPaths_Beg))
	for _, path := range fc.BlockedPaths_Beg {
		hash = hash*31 + uint64(len(path))
		for i := 0; i < len(path); i++ {
			hash = hash*31 + uint64(path[i])
		}
	}

	hash = hash*31 + uint64(len(fc.EntryPoints))
	for _, entryPoint := range fc.EntryPoints {
		hash = hash*31 + uint64(len(entryPoint))
		for i := 0; i < len(entryPoint); i++ {
			hash = hash*31 + uint64(entryPoint[i])
		}
	}

	hash = hash*31 + uint64(len(fc.PathPrefix))
	for i := 0; i < len(fc.PathPrefix); i++ {
		hash = hash*31 + uint64(fc.PathPrefix[i])
	}

	return hash
}
