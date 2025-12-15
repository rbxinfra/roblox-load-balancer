package consul

import (
	capi "github.com/hashicorp/consul/api"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
)

var apiClient *capi.Client

// InitializeConsul initializes the Consul API client
// with the default configuration or based on the Daemon
// configuration.
func InitializeConsul(config *configuration.Config) error {
	consulConfig := capi.DefaultConfig()

	if config.Consul != nil {
		daemonConsulConfig := config.Consul

		if daemonConsulConfig.Address != "" {
			consulConfig.Address = daemonConsulConfig.Address
		}

		if daemonConsulConfig.Scheme != "" {
			consulConfig.Scheme = daemonConsulConfig.Scheme
		}

		if daemonConsulConfig.PathPrefix != "" {
			consulConfig.PathPrefix = daemonConsulConfig.PathPrefix
		}

		if daemonConsulConfig.Datacenter != "" {
			consulConfig.Datacenter = daemonConsulConfig.Datacenter
		}

		if daemonConsulConfig.HttpAuth != nil {
			consulConfig.HttpAuth = &capi.HttpBasicAuth{
				Username: daemonConsulConfig.HttpAuth.Username,
				Password: daemonConsulConfig.HttpAuth.Password,
			}
		}

		if daemonConsulConfig.WaitTime != 0 {
			consulConfig.WaitTime = daemonConsulConfig.WaitTime
		}

		if daemonConsulConfig.Token != "" {
			consulConfig.Token = daemonConsulConfig.Token
		}

		if daemonConsulConfig.TokenFile != "" {
			consulConfig.TokenFile = daemonConsulConfig.TokenFile
		}

		if daemonConsulConfig.Namespace != "" {
			consulConfig.Namespace = daemonConsulConfig.Namespace
		}

		if daemonConsulConfig.Partition != "" {
			consulConfig.Partition = daemonConsulConfig.Partition
		}

		if daemonConsulConfig.TLSConfig != nil {
			consulConfig.TLSConfig = capi.TLSConfig(*daemonConsulConfig.TLSConfig)
		}
	}

	client, err := capi.NewClient(consulConfig)
	if err != nil {
		return err
	}

	apiClient = client

	return nil
}

// GetClient gets the current Consul API Client.
func GetClient() *capi.Client {
	return apiClient
}
