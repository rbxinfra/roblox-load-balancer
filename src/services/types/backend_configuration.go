package types

// BackendConfiguration represents the configuration
// for the backend of a HAProxy service (such as load balancing
// or health checks).
type BackendConfiguration struct {
	// Balance is the load balancing mode.
	//
	// Defaults to roundrobin
	Balance string

	// HashType is the type of hash to use.
	//
	// Defaults to consistent
	HashType string

	// BlockedPaths is a list of directly blocked paths.
	BlockedPaths []string

	// BlockedPaths_Beg is a list of blocked path prefixes.
	BlockedPaths_Beg []string

	// Del_Headers is a list of headers to delete from the request.
	Del_Headers []string

	// SetHostHeader sets the host header to the specified value.
	SetHostHeader string
}

// Hash computes a hash of the BackendConfiguration
func (bc *BackendConfiguration) Hash() uint64 {
	var hash uint64 = 17

	hash = hash*31 + uint64(len(bc.Balance))
	for i := 0; i < len(bc.Balance); i++ {
		hash = hash*31 + uint64(bc.Balance[i])
	}

	hash = hash*31 + uint64(len(bc.HashType))
	for i := 0; i < len(bc.HashType); i++ {
		hash = hash*31 + uint64(bc.HashType[i])
	}

	hash = hash*31 + uint64(len(bc.BlockedPaths))
	for _, path := range bc.BlockedPaths {
		hash = hash*31 + uint64(len(path))
		for i := 0; i < len(path); i++ {
			hash = hash*31 + uint64(path[i])
		}
	}

	hash = hash*31 + uint64(len(bc.BlockedPaths_Beg))
	for _, path := range bc.BlockedPaths_Beg {
		hash = hash*31 + uint64(len(path))
		for i := 0; i < len(path); i++ {
			hash = hash*31 + uint64(path[i])
		}
	}

	hash = hash*31 + uint64(len(bc.Del_Headers))
	for _, header := range bc.Del_Headers {
		hash = hash*31 + uint64(len(header))
		for i := 0; i < len(header); i++ {
			hash = hash*31 + uint64(header[i])
		}
	}

	hash = hash*31 + uint64(len(bc.SetHostHeader))
	for i := 0; i < len(bc.SetHostHeader); i++ {
		hash = hash*31 + uint64(bc.SetHostHeader[i])
	}

	return hash
}
