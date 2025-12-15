package types

// ServiceNode represent a node within a service.
type ServiceNode struct {
	// Name is the name of this node.
	Name string

	// Address is the address of this node.
	Address string

	// Port is the port of this node.
	Port int
}

// Hash computes a hash of the ServiceNode
func (sn *ServiceNode) Hash() uint64 {
	var hash uint64 = 17

	hash = hash*31 + uint64(len(sn.Name))
	for i := 0; i < len(sn.Name); i++ {
		hash = hash*31 + uint64(sn.Name[i])
	}

	hash = hash*31 + uint64(len(sn.Address))
	for i := 0; i < len(sn.Address); i++ {
		hash = hash*31 + uint64(sn.Address[i])
	}

	hash = hash*31 + uint64(sn.Port)

	return hash
}
