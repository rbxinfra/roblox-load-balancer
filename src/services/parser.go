package services

import (
	"fmt"
	"strings"

	capi "github.com/hashicorp/consul/api"
	"github.com/traefik/paerser/parser"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/services/types"
)

const (
	PROTO_HTTP  = "http"
	PROTO_HTTPS = "https"
	PROTO_H2C   = "h2c"

	ALG_RR          = "roundrobin"
	HASH_CONSISTENT = "consistent"
)

func validateLabelsConfig(config *types.ServiceConfig, entryPoints map[string]*configuration.EntrypointConfig) error {
	if config.Protocol == "" {
		config.Protocol = PROTO_HTTP
	}

	if !strings.EqualFold(config.Protocol, PROTO_HTTP) &&
		!strings.EqualFold(config.Protocol, PROTO_HTTPS) &&
		!strings.EqualFold(config.Protocol, PROTO_H2C) {
		return fmt.Errorf("Invalid protocol specified, expected one of http, https, or h2c, got %s", config.Protocol)
	}

	if config.Be == nil {
		config.Be = &types.BackendConfiguration{}
	}

	if config.Fe == nil {
		config.Fe = &types.FrontendConfiguration{}
	}

	if config.Be.Balance == "" {
		config.Be.Balance = ALG_RR
	}

	if config.Be.HashType == "" {
		config.Be.HashType = HASH_CONSISTENT
	}

	if len(config.Fe.Fqdn) == 0 {
		return fmt.Errorf("Service must specify at least one FQDN.")
	}

	if len(config.Fe.EntryPoints) == 0 {
		for entryPoint := range entryPoints {
			config.Fe.EntryPoints = append(config.Fe.EntryPoints, entryPoint)
		}
	}

	for _, entryPoint := range config.Fe.EntryPoints {
		if _, ok := entryPoints[entryPoint]; !ok {
			return fmt.Errorf("Unknown entrypoint %s", entryPoint)
		}
	}

	return nil
}

// ParseServicesFromConsul parses services from Consul.
func ParseServicesFromConsul(serviceNodes map[string][]*capi.CatalogService, config *configuration.Config) ([]*types.Service, error) {
	services := make([]*types.Service, 0, len(serviceNodes))

	for serviceName, serviceInstances := range serviceNodes {
		service, err := parseServiceFromConsul(serviceName, serviceInstances, config)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}

func parseServiceFromConsul(serviceName string, serviceInstances []*capi.CatalogService, config *configuration.Config) (*types.Service, error) {
	service := &types.Service{
		ServiceName: serviceName,
		Nodes:       make([]*types.ServiceNode, 0, len(serviceInstances)),
		Config:      &types.ServiceConfig{},
	}

	for _, entry := range serviceInstances {
		serviceNode := &types.ServiceNode{
			Name:    entry.Node,
			Address: entry.Address,
			Port:    entry.ServicePort,
		}

		if externalSource, ok := entry.ServiceMeta["external-source"]; ok && externalSource == "nomad" {
			serviceNode.Name = strings.Split(entry.ServiceID, "-")[2] // The short ALLOC id
		}

		service.Nodes = append(service.Nodes, serviceNode)
	}

	labels := tagsToNeutralLabels(serviceInstances[0].ServiceTags, config.Prefix)
	if err := parser.Decode(labels, service.Config, "haproxy"); err != nil {
		return nil, err
	}

	if err := validateLabelsConfig(service.Config, config.Entrypoints); err != nil {
		return nil, err
	}

	return service, nil
}
