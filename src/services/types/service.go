package types

// Service represents a service.
type Service struct {
	// ServiceName is the name of the service.
	ServiceName string

	// Config is the label config from Consul.
	Config *ServiceConfig

	// Nodes is the nodes of this service.
	Nodes []*ServiceNode
}

// Hash computes a hash of the Service
func (s *Service) Hash() uint64 {
	var hash uint64 = 17

	hash = hash*31 + uint64(len(s.ServiceName))
	for i := 0; i < len(s.ServiceName); i++ {
		hash = hash*31 + uint64(s.ServiceName[i])
	}

	if s.Config != nil {
		hash = hash*31 + s.Config.Hash()
	}

	for _, node := range s.Nodes {
		hash = hash*31 + node.Hash()
	}

	return hash
}
