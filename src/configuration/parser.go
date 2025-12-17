package configuration

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
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

func applyDefaultsAndValidateConfiguration(config *Config) error {
	if config.Prefix == "" {
		config.Prefix = DefaultLabelPrefix
	}

	if config.TemplateFilePath == "" {
		return fmt.Errorf("config.TemplateFilePath must be specified!")
	}

	if !filepath.IsAbs(config.TemplateFilePath) {
		absPath, err := filepath.Abs(config.TemplateFilePath)
		if err != nil {
			return err
		}
		config.TemplateFilePath = absPath
	}

	if _, err := os.Stat(config.TemplateFilePath); err != nil {
		return err
	}

	if config.OutputFilePath == "" {
		config.OutputFilePath = DefaultOutputFilePath
	}

	if !filepath.IsAbs(config.OutputFilePath) {
		absPath, err := filepath.Abs(config.OutputFilePath)
		if err != nil {
			return err
		}
		config.OutputFilePath = absPath
	}

	if config.RefreshInterval == nil {
		config.RefreshInterval = new(time.Duration)
		*config.RefreshInterval = time.Minute * 5
	}

	if config.HAProxy == nil {
		config.HAProxy = new(HAProxyConfig)
	}

	if config.HAProxy.Path == "" {
		haproxyPath, err := exec.LookPath("haproxy")
		if err != nil {
			return err
		}

		config.HAProxy.Path = haproxyPath
	}

	if len(config.HAProxy.Args) == 0 {
		config.HAProxy.Args = append(config.HAProxy.Args, "-W", "-db", "-f", config.OutputFilePath)
	}

	if config.HAProxy.MaxStartAttempts == nil {
		config.HAProxy.MaxStartAttempts = new(int)
		*config.HAProxy.MaxStartAttempts = 3
	}

	if config.HAProxy.StdoutLogFilePath == "" {
		config.HAProxy.StdoutLogFilePath = "/var/log/haproxy/stdout"
	}

	if config.HAProxy.StderrLogFilePath == "" {
		config.HAProxy.StderrLogFilePath = "/var/log/haproxy/stderr"
	}

	if config.ServersConfig == nil {
		config.ServersConfig = new(ServersConfig)
	}

	if config.ServersConfig.Default == nil {
		config.ServersConfig.Default = new(ServerConfig)
		config.ServersConfig.Default.Interval = time.Second * 5
		config.ServersConfig.Default.Fall = 3
		config.ServersConfig.Default.Rise = 2
	}

	if config.ServersConfig.PerServer == nil {
		config.ServersConfig.PerServer = new(ServerConfig)
		config.ServersConfig.PerServer.Interval = time.Second * 10
		config.ServersConfig.PerServer.Fall = 1
		config.ServersConfig.PerServer.Rise = 1
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
					Type:  "status",
					Match: true,
					Value: "200",
				})
			}
		}
	}

	if len(config.Entrypoints) == 0 {
		return fmt.Errorf("config.Entrypoints must have at least one entry!")
	}

	return nil
}

// ParseConfiguration parses the configuration from Json, Yaml, or TOML.
// It then sets defaults and validates the configuration.
func ParseConfiguration() (*Config, error) {
	if *flags.ConfigurationFilePath == "" {
		return nil, fmt.Errorf("The configuration file must be specified!")
	}

	glog.Infof("Loading configuration file from %s...", *flags.ConfigurationFilePath)

	config, err := parseFileDependingOnExtension(*flags.ConfigurationFilePath)
	if err != nil {
		return nil, err
	}

	err = applyDefaultsAndValidateConfiguration(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
