package services

import (
	"context"
	"fmt"

	capi "github.com/hashicorp/consul/api"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/consul"
)

var gLastConsulIndex uint64 = 0

// FetchLatestServices fetches a map of service name to service instance from Consul.
// This method will block until an update is made to the Catalog.
func FetchLatestServices(ctx context.Context, config *configuration.Config) (map[string][]*capi.CatalogService, error) {
	catalog := consul.GetClient().Catalog()

	options := capi.QueryOptions{
		WaitIndex: gLastConsulIndex,
		Filter: fmt.Sprintf("\"%s.enable=true\" in ServiceTags", config.Prefix),
	}

	services, meta, err := catalog.Services(options.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	options = capi.QueryOptions{}

	gLastConsulIndex = meta.LastIndex

	result := make(map[string][]*capi.CatalogService)
	
	for service := range services {
		serviceNodes, _, err := catalog.Service(service, fmt.Sprintf("%s.enable=true", config.Prefix), options.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		result[service] = serviceNodes
	}

	return result, nil
}
