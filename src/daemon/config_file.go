package daemon

import (
	"context"
	"os"

	"github.com/golang/glog"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/haproxy"
	"github.rbx.com/roblox/roblox-load-balancer/services"
	"github.rbx.com/roblox/roblox-load-balancer/services/types"
)

// UpdateHAProxyConfigurationFile updates the HAProxy configuration file
// from Consul and returns the current services.
func UpdateHAProxyConfigurationFile(ctx context.Context, config *configuration.Config) ([]*types.Service, error) {
	glog.V(100).Infoln("Trying to update HAProxy configuration from Consul.")

	consulSvcs, err := services.FetchLatestServices(ctx, config)
	if err != nil {
		return nil, err
	}

	if len(consulSvcs) == 0 {
		glog.V(100).Infoln("Consul returned an empty list of services, no backends will be routed!")
	}

	svcs, err := services.ParseServicesFromConsul(consulSvcs, config)
	if err != nil {
		return nil, err
	}

	backendsMap := services.BuildBackends(svcs, config)
	rulesMap := services.BuildRules(svcs, config)

	parsedFile, err := haproxy.BuildTemplateFile(backendsMap, rulesMap, config)
	if err != nil {
		return nil, err
	}

	glog.V(100).Infof("Writing parsed HAProxy configuration file to %s", config.OutputFilePath)

	outputFile, err := os.OpenFile(config.OutputFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(parsedFile)
	if err != nil {
		return nil, err
	}

	return svcs, nil
}
