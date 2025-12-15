package services

import (
	"fmt"
	"slices"
	"strings"

	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/services/types"
)

// BuildBackends builds a map of entrypoint to backends.
func BuildBackends(services []*types.Service, config *configuration.Config) map[string]string {
	entrypointMap := make(map[string]string)

	for entryPoint := range config.Entrypoints {
		var backends string = ""

		for _, service := range services {
			backends += buildBackendForEntrypoint(entryPoint, service, config)
			backends += "\n"
		}

		entrypointMap[entryPoint] = backends
	}

	return entrypointMap
}

// BuildRules builds a map of entrypoint to frontend rules.
func BuildRules(services []*types.Service, config *configuration.Config) map[string]string {
	entrypointMap := make(map[string]string)

	for entryPoint := range config.Entrypoints {
		var rules string = ""

		for _, service := range services {
			rules += buildRuleForEntrypoint(entryPoint, service)
			rules += "\n"
		}

		entrypointMap[entryPoint] = rules
	}

	return entrypointMap

}

func buildRuleForEntrypoint(entryPoint string, service *types.Service) string {
	if !slices.Contains(service.Config.Fe.EntryPoints, entryPoint) {
		return ""
	}

	var result string = fmt.Sprintf("  use_backend %s.%s if { hdr(host) -i %s }", service.ServiceName, entryPoint, strings.Join(service.Config.Fe.Fqdn, " "))

	if len(service.Config.Fe.BlockedPaths) > 0 {
		result += fmt.Sprintf(" !{ path %s }", strings.Join(service.Config.Fe.BlockedPaths, " "))
	}

	if len(service.Config.Fe.BlockedPaths_Beg) > 0 {
		result += fmt.Sprintf(" !{ path_beg %s }", strings.Join(service.Config.Fe.BlockedPaths_Beg, " "))
	}

	if service.Config.Fe.PathPrefix != "" {
		result += fmt.Sprintf(" { path_beg %s }", service.Config.Fe.PathPrefix)
	}

	return result
}

func buildBackendForEntrypoint(entryPoint string, service *types.Service, config *configuration.Config) string {
	if !slices.Contains(service.Config.Fe.EntryPoints, entryPoint) {
		return ""
	}

	var result string = fmt.Sprintf("backend %s.%s\n", service.ServiceName, entryPoint)

	result += "  default-server pool-purge-delay 30s\n"

	if entryPointConfig, ok := config.Entrypoints[entryPoint]; ok {
		result += entryPointConfig.String()
	}

	if len(service.Config.Be.BlockedPaths) > 0 {
		result += fmt.Sprintf("  http-request deny if { path %s }\n", strings.Join(service.Config.Fe.BlockedPaths, " "))
	}

	if len(service.Config.Be.BlockedPaths_Beg) > 0 {
		result += fmt.Sprintf("  http-request deny if { path_beg %s }\n", strings.Join(service.Config.Fe.BlockedPaths_Beg, " "))
	}

	if len(service.Config.Be.Del_Headers) > 0 {
		for _, header := range service.Config.Be.Del_Headers {
			result += fmt.Sprintf("  http-request del-header %s\n", header)
		}
	}

	if service.Config.Fe.PathPrefix != "" {
		result += fmt.Sprintf("  http-request replace-path %s(/)?(.*) /\\2\n", service.Config.Fe.PathPrefix)
	}

	if service.Config.Be.SetHostHeader != "" {
		result += fmt.Sprintf("  http-request set-header Host %s\n", service.Config.Be.SetHostHeader)
	}

	result += fmt.Sprintf("  balance %s\n", service.Config.Be.Balance)
	result += fmt.Sprintf("  hash-type %s\n", service.Config.Be.HashType)

	var healthCheck *configuration.HealthCheckConfig

	if serviceHealthCheck, ok := config.HealthChecks[service.ServiceName]; ok {
		healthCheck = serviceHealthCheck.Copy()
	} else if defaultHealthCheck, ok := config.HealthChecks["default"]; ok {
		healthCheck = defaultHealthCheck.Copy()

		config.HealthChecks[service.ServiceName] = healthCheck
	}

	if healthCheck != nil {
		if len(healthCheck.Send) == 0 {
			healthCheck.Send = append(healthCheck.Send, configuration.HealthCheckSend{
				Headers: map[string]string{"host": service.Config.Fe.Fqdn[0]},
			})
		}

		result += healthCheck.String()
	}

	result += fmt.Sprintf("  default-server inter %s rise %d fall %d\n", config.ServersConfig.Default.Interval, config.ServersConfig.Default.Rise, config.ServersConfig.PerServer.Fall)

	for _, node := range service.Nodes {
		result += fmt.Sprintf("  server %s %s:%d", node.Name, node.Address, node.Port)

		if service.Config.Protocol == PROTO_HTTPS {
			if config.TLSBundleFilePath != "" {
				result += fmt.Sprintf(" ssl verify required ca-file %s", config.TLSBundleFilePath)
			} else {
				result += " ssl verify none"
			}
		}

		result += fmt.Sprintf(" check inter %s rise %d fall %d", config.ServersConfig.PerServer.Interval, config.ServersConfig.PerServer.Rise, config.ServersConfig.PerServer.Fall)

		if service.Config.Protocol == PROTO_H2C {
			result += " proto h2"
		}

		result += "\n"
	}

	return result
}
