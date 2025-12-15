package configuration

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/BurntSushi/toml"
	"github.rbx.com/roblox/roblox-load-balancer/flags"
	"gopkg.in/yaml.v3"
)

const (
	DefaultLabelPrefix    = "haproxy"
	DefaultOutputFilePath = "/usr/local/etc/haproxy/haproxy.cfg"
)

func parseYAMLFile(fileName string) (*Config, error) {
        var conf Config

        yamlFile, err := os.Open(fileName)
        if err != nil {
                return nil, err
        }

        defer yamlFile.Close()

        yamlParser := yaml.NewDecoder(yamlFile)
        if err = yamlParser.Decode(&conf); err != nil {
                return nil, err
        }

        return &conf, nil
}

func parseTOMLFile(fileName string) (*Config, error) {
        var conf Config

        tomlFile, err := os.Open(fileName)
        if err != nil {
                return nil, err
        }

        defer tomlFile.Close()

        tomlParser := toml.NewDecoder(tomlFile)
        if _, err = tomlParser.Decode(&conf); err != nil {
                return nil, err
        }

        return &conf, nil
}

func parseFileDependingOnExtension(fileName string) (*Config, error) {
        if fileName == "" {
                return nil, nil
        }

        fileExtension := path.Ext(fileName)

        switch fileExtension {
        case ".yml", ".yaml", ".json":
                return parseYAMLFile(fileName)
        case ".toml":
                return parseTOMLFile(fileName)
        default:
                return nil, nil
        }
}

func applyDefaults(config *Config) {
	if config.Prefix == "" {
		config.Prefix = DefaultLabelPrefix
	}

	if config.OutputFilePath == "" {
		config.OutputFilePath = DefaultOutputFilePath
	}

	if config.HAProxy == nil {
		config.HAProxy = &HAProxyConfig{
			Path: "haproxy",
			Args: []string{"-W", "-db", "-f", config.OutputFilePath},
		}
	}

	if config.ServersConfig == nil {
		config.ServersConfig = &ServersConfig{
			Default: &ServerConfig{ Interval: time.Second*5, Fall: 3, Rise: 2 },
			PerServer: &ServerConfig{ Interval: time.Second*10, Fall: 1, Rise: 1 },
		}
	}

	if len(config.HealthChecks) > 0 {
		for _, check := range config.HealthChecks {
			if check.Option.Method == "" {
				check.Option.Method = "HEAD"
			}

			if check.Option.Version == "" {
				check.Option.Version = "HTTP/1.1"
			}

			if len(check.Expect) == 0 {
				check.Expect = append(check.Expect, HealthCheckExpect{
					Type: "status",
					Match: true,
					Value: "200",
				})
			}
		}
	}
}

// ParseConfiguration parses the configuration from Json, Yaml, or TOML.
// It then sets defaults and validates the configuration.
func ParseConfiguration() (*Config, error) {
	if *flags.ConfigurationFilePath == "" {
		return nil, fmt.Errorf("The configuration file must be specified!")
	}

	config, err := parseFileDependingOnExtension(*flags.ConfigurationFilePath)
	if err != nil {
		return nil, err
	}

	applyDefaults(config)
	if config.TemplateFilePath == "" {
		return nil, fmt.Errorf("config.TemplateFilePath must be specified!")
	}

	if _, err := os.Stat(config.TemplateFilePath); err != nil {
		return nil, err
	}

	if len(config.Entrypoints) == 0 {
		return nil, fmt.Errorf("config.Entrypoints must have at least one entry!")
	}

	return config, nil
}
